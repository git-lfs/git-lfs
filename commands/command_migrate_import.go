package commands

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	// Only perform `git-checkout(1) -f` if the repository is
	// non-bare.
	if bare, _ := git.IsBare(); !bare {
		t := l.Waiter("migrate: checkout")
		err := git.Checkout("", nil, true)
		t.Complete()

		if err != nil {
			ExitWithError(err)
		}
	}
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
