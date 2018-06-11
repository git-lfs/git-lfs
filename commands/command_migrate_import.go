package commands

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/git/githistory"
	"github.com/git-lfs/git-lfs/git/odb"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/tasklog"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/spf13/cobra"
)

func migrateImportCommand(cmd *cobra.Command, args []string) {
	l := tasklog.NewLogger(os.Stderr)
	defer l.Close()

	db, err := getObjectDatabase()
	if err != nil {
		ExitWithError(err)
	}
	defer db.Close()

	if migrateNoRewrite {
		if len(args) == 0 {
			ExitWithError(errors.Errorf("fatal: expected one or more files with --no-rewrite"))
		}

		ref, err := git.CurrentRef()
		if err != nil {
			ExitWithError(errors.Wrap(err, "fatal: unable to find current reference"))
		}

		sha, _ := hex.DecodeString(ref.Sha)
		commit, err := db.Commit(sha)
		if err != nil {
			ExitWithError(errors.Wrap(err, "fatal: unable to load commit"))
		}

		root := commit.TreeID

		filter := git.GetAttributeFilter(cfg.LocalWorkingDir(), cfg.LocalGitDir())
		if len(filter.Include()) == 0 {
			ExitWithError(errors.Errorf("fatal: no Git LFS filters found in .gitattributes"))
		}

		gf := lfs.NewGitFilter(cfg)

		for _, file := range args {
			if !filter.Allows(file) {
				ExitWithError(errors.Errorf("fatal: file %s did not match any Git LFS filters in .gitattributes", file))
			}
		}

		for _, file := range args {
			root, err = rewriteTree(gf, db, root, file)
			if err != nil {
				ExitWithError(errors.Wrapf(err, "fatal: could not rewrite %q", file))
			}
		}

		name, email := cfg.CurrentCommitter()
		author := fmt.Sprintf("%s <%s>", name, email)

		oid, err := db.WriteCommit(&odb.Commit{
			Author:    author,
			Committer: author,
			ParentIDs: [][]byte{sha},
			Message:   generateMigrateCommitMessage(cmd, strings.Join(args, ",")),
			TreeID:    root,
		})

		if err != nil {
			ExitWithError(errors.Wrap(err, "fatal: unable to write commit"))
		}

		if err := git.UpdateRef(ref, oid, "git lfs migrate import --no-rewrite"); err != nil {
			ExitWithError(errors.Wrap(err, "fatal: unable to update ref"))
		}

		if err := checkoutNonBare(l); err != nil {
			ExitWithError(errors.Wrap(err, "fatal: could not checkout"))
		}

		return
	}

	rewriter := getHistoryRewriter(cmd, db, l)

	tracked := trackedFromFilter(rewriter.Filter())
	exts := tools.NewOrderedSet()
	gitfilter := lfs.NewGitFilter(cfg)

	migrate(args, rewriter, l, &githistory.RewriteOptions{
		Verbose:           migrateVerbose,
		ObjectMapFilePath: objectMapFilePath,
		BlobFn: func(path string, b *odb.Blob) (*odb.Blob, error) {
			if filepath.Base(path) == ".gitattributes" {
				return b, nil
			}

			var buf bytes.Buffer

			if _, err := clean(gitfilter, &buf, b.Contents, path, b.Size); err != nil {
				return nil, err
			}

			if ext := filepath.Ext(path); len(ext) > 0 {
				exts.Add(fmt.Sprintf("*%s filter=lfs diff=lfs merge=lfs -text", ext))
			}

			return &odb.Blob{
				Contents: &buf, Size: int64(buf.Len()),
			}, nil
		},

		TreeCallbackFn: func(path string, t *odb.Tree) (*odb.Tree, error) {
			if path != "/" {
				// Ignore non-root trees.
				return t, nil
			}

			ours := tracked
			if ours.Cardinality() == 0 {
				// If there were no explicitly tracked
				// --include, --exclude filters, assume that the
				// include set is the wildcard filepath
				// extensions of files tracked.
				ours = exts

				if ours.Cardinality() == 0 {
					// If it is still the case that we have
					// no patterns to track, that means that
					// we are in a tree that does not
					// require .gitattributes changes.
					//
					// We can return early to avoid
					// comparing and saving an identical
					// tree.
					return t, nil
				}
			}

			theirs, err := trackedFromAttrs(db, t)
			if err != nil {
				return nil, err
			}

			// Create a blob of the attributes that are optionally
			// present in the "t" tree's .gitattributes blob, and
			// union in the patterns that we've tracked.
			//
			// Perform this Union() operation each time we visit a
			// root tree such that if the underlying .gitattributes
			// is present and has a diff between commits in the
			// range of commits to migrate, those changes are
			// preserved.
			blob, err := trackedToBlob(db, theirs.Clone().Union(ours))
			if err != nil {
				return nil, err
			}

			// Finally, return a copy of the tree "t" that has the
			// new .gitattributes file included/replaced.
			return t.Merge(&odb.TreeEntry{
				Name:     ".gitattributes",
				Filemode: 0100644,
				Oid:      blob,
			}), nil
		},

		UpdateRefs: true,
	})

	if err := checkoutNonBare(l); err != nil {
		ExitWithError(errors.Wrap(err, "fatal: could not checkout"))
	}
}

// generateMigrateCommitMessage generates a commit message used with
// --no-rewrite, using --message (if given) or generating one if it isn't.
func generateMigrateCommitMessage(cmd *cobra.Command, patterns string) string {
	if cmd.Flag("message").Changed {
		return migrateCommitMessage
	}
	return fmt.Sprintf("%s: convert to Git LFS", patterns)
}

// checkoutNonBare forces a checkout of the current reference, so long as the
// repository is non-bare.
//
// It returns nil on success, and a non-nil error on failure.
func checkoutNonBare(l *tasklog.Logger) error {
	if bare, _ := git.IsBare(); bare {
		return nil
	}

	t := l.Waiter("migrate: checkout")
	defer t.Complete()

	return git.Checkout("", nil, true)
}

// trackedFromFilter returns an ordered set of strings where each entry is a
// line in the .gitattributes file. It adds/removes the fiter/diff/merge=lfs
// attributes based on patterns included/excldued in the given filter.
func trackedFromFilter(filter *filepathfilter.Filter) *tools.OrderedSet {
	tracked := tools.NewOrderedSet()

	for _, include := range filter.Include() {
		tracked.Add(fmt.Sprintf("%s filter=lfs diff=lfs merge=lfs -text", escapeAttrPattern(include)))
	}

	for _, exclude := range filter.Exclude() {
		tracked.Add(fmt.Sprintf("%s text -filter -merge -diff", escapeAttrPattern(exclude)))
	}

	return tracked
}

var (
	// attrsCache maintains a cache from the hex-encoded SHA1 of a
	// .gitattributes blob to the set of patterns parsed from that blob.
	attrsCache = make(map[string]*tools.OrderedSet)
)

// trackedFromAttrs returns an ordered line-delimited set of the contents of a
// .gitattributes blob in a given tree "t".
//
// It returns an empty set if no attributes file could be found, or an error if
// it could not otherwise be opened.
func trackedFromAttrs(db *odb.ObjectDatabase, t *odb.Tree) (*tools.OrderedSet, error) {
	var oid []byte

	for _, e := range t.Entries {
		if strings.ToLower(e.Name) == ".gitattributes" && e.Type() == odb.BlobObjectType {
			oid = e.Oid
			break
		}
	}

	if oid == nil {
		// TODO(@ttaylorr): make (*tools.OrderedSet)(nil) a valid
		// receiver for non-mutative methods.
		return tools.NewOrderedSet(), nil
	}

	sha1 := hex.EncodeToString(oid)

	if s, ok := attrsCache[sha1]; ok {
		return s, nil
	}

	blob, err := db.Blob(oid)
	if err != nil {
		return nil, err
	}

	attrs := tools.NewOrderedSet()

	scanner := bufio.NewScanner(blob.Contents)
	for scanner.Scan() {
		attrs.Add(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	attrsCache[sha1] = attrs

	return attrsCache[sha1], nil
}

// trackedToBlob writes and returns the OID of a .gitattributes blob based on
// the patterns given in the ordered set of patterns, "patterns".
func trackedToBlob(db *odb.ObjectDatabase, patterns *tools.OrderedSet) ([]byte, error) {
	var attrs bytes.Buffer

	for pattern := range patterns.Iter() {
		fmt.Fprintf(&attrs, "%s\n", pattern)
	}

	return db.WriteBlob(&odb.Blob{
		Contents: &attrs,
		Size:     int64(attrs.Len()),
	})
}

// rewriteTree replaces the blob at the provided path within the given tree with
// a git lfs pointer. It will recursively rewrite any subtrees along the path to the
// blob.
func rewriteTree(gf *lfs.GitFilter, db *odb.ObjectDatabase, root []byte, path string) ([]byte, error) {
	tree, err := db.Tree(root)
	if err != nil {
		return nil, err
	}

	splits := strings.SplitN(path, "/", 2)

	switch len(splits) {
	case 1:
		// The path points to an entry at the root of this tree, so it must be a blob.
		// Try to replace this blob with a Git LFS pointer.
		index := findEntry(tree, splits[0])
		if index < 0 {
			return nil, errors.Errorf("unable to find entry %s in tree", splits[0])
		}

		blobEntry := tree.Entries[index]
		blob, err := db.Blob(blobEntry.Oid)
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer

		if _, err := clean(gf, &buf, blob.Contents, blobEntry.Name, blob.Size); err != nil {
			return nil, err
		}

		newOid, err := db.WriteBlob(&odb.Blob{
			Contents: &buf,
			Size:     int64(buf.Len()),
		})

		if err != nil {
			return nil, err
		}

		tree = tree.Merge(&odb.TreeEntry{
			Name:     splits[0],
			Filemode: blobEntry.Filemode,
			Oid:      newOid,
		})
		return db.WriteTree(tree)

	case 2:
		// The path points to an entry in a subtree contained at the root of the tree.
		// Recursively rewrite the subtree.
		head, tail := splits[0], splits[1]

		index := findEntry(tree, head)
		if index < 0 {
			return nil, errors.Errorf("unable to find entry %s in tree", head)
		}

		subtreeEntry := tree.Entries[index]
		if subtreeEntry.Type() != odb.TreeObjectType {
			return nil, errors.Errorf("migrate: expected %s to be a tree, got %s", head, subtreeEntry.Type())
		}

		rewrittenSubtree, err := rewriteTree(gf, db, subtreeEntry.Oid, tail)
		if err != nil {
			return nil, err
		}

		tree = tree.Merge(&odb.TreeEntry{
			Filemode: subtreeEntry.Filemode,
			Name:     subtreeEntry.Name,
			Oid:      rewrittenSubtree,
		})

		return db.WriteTree(tree)

	default:
		return nil, errors.Errorf("error parsing path %s", path)
	}
}

// findEntry searches a tree for the desired entry, and returns the index of that
// entry within the tree's Entries array
func findEntry(t *odb.Tree, name string) int {
	for i, entry := range t.Entries {
		if entry.Name == name {
			return i
		}
	}

	return -1
}
