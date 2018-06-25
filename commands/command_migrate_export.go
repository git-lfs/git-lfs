package commands

import (
	"fmt"
	"os"
	"path/filepath"

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

func migrateExportCommand(cmd *cobra.Command, args []string) {
	l := tasklog.NewLogger(os.Stderr)
	defer l.Close()

	db, err := getObjectDatabase()
	if err != nil {
		ExitWithError(err)
	}
	defer db.Close()

	rewriter := getHistoryRewriter(cmd, db, l)

	filter := rewriter.Filter()
	if len(filter.Include()) <= 0 {
		ExitWithError(errors.Errorf("fatal: one or more files must be specified with --include"))
	}

	tracked := trackedFromExportFilter(filter)
	gitfilter := lfs.NewGitFilter(cfg)

	opts := &githistory.RewriteOptions{
		Verbose:           migrateVerbose,
		ObjectMapFilePath: objectMapFilePath,
		BlobFn: func(path string, b *odb.Blob) (*odb.Blob, error) {
			if filepath.Base(path) == ".gitattributes" {
				return b, nil
			}

			ptr, err := lfs.DecodePointer(b.Contents)
			if errors.IsNotAPointerError(err) {
				return b, nil
			}
			if err != nil {
				return nil, err
			}

			downloadPath, err := gitfilter.ObjectPath(ptr.Oid)
			if err != nil {
				return nil, err
			}

			return odb.NewBlobFromFile(downloadPath)
		},

		TreeCallbackFn: func(path string, t *odb.Tree) (*odb.Tree, error) {
			if path != "/" {
				// Ignore non-root trees.
				return t, nil
			}

			ours := tracked
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
	}

	requireInRepo()

	opts, err = rewriteOptions(args, opts, l)
	if err != nil {
		ExitWithError(err)
	}

	// If we have a valid remote, pre-download all objects using the Transfer Queue
	if remoteURL := getAPIClient().Endpoints.RemoteEndpoint("download", cfg.Remote()).Url; remoteURL != "" {
		q := newDownloadQueue(getTransferManifestOperationRemote("Download", cfg.Remote()), cfg.Remote())
		if err := rewriter.ScanForPointers(q, opts, gitfilter); err != nil {
			ExitWithError(err)
		}

		q.Wait()

		for _, err := range q.Errors() {
			if err != nil {
				ExitWithError(err)
			}
		}
	}

	// Perform the rewrite
	if _, err := rewriter.Rewrite(opts); err != nil {
		ExitWithError(err)
	}

	// Only perform `git-checkout(1) -f` if the repository is non-bare.
	if bare, _ := git.IsBare(); !bare {
		t := l.Waiter("migrate: checkout")
		err := git.Checkout("", nil, true)
		t.Complete()

		if err != nil {
			ExitWithError(err)
		}
	}
}

// trackedFromExportFilter returns an ordered set of strings where each entry
// is a line we intend to place in the .gitattributes file. It adds/removes the
// filter/diff/merge=lfs attributes based on patterns included/excluded in the
// given filter. Since `migrate export` removes files from Git LFS, it will
// remove attributes for included files, and add attributes for excluded files
func trackedFromExportFilter(filter *filepathfilter.Filter) *tools.OrderedSet {
	tracked := tools.NewOrderedSet()

	for _, include := range filter.Include() {
		tracked.Add(fmt.Sprintf("%s text !filter !merge !diff", escapeAttrPattern(include)))
	}

	for _, exclude := range filter.Exclude() {
		tracked.Add(fmt.Sprintf("%s filter=lfs diff=lfs merge=lfs -text", escapeAttrPattern(exclude)))
	}

	return tracked
}
