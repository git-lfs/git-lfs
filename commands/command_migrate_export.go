package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/filepathfilter"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/git/githistory"
	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/tasklog"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/git-lfs/gitobj/v2"
	"github.com/spf13/cobra"
)

func migrateExportCommand(cmd *cobra.Command, args []string) {
	ensureWorkingCopyClean(os.Stdin, os.Stderr)

	l := tasklog.NewLogger(os.Stderr,
		tasklog.ForceProgress(cfg.ForceProgress()),
	)
	defer l.Close()

	db, err := getObjectDatabase()
	if err != nil {
		ExitWithError(err)
	}
	defer db.Close()

	rewriter := getHistoryRewriter(cmd, db, l)

	filter := rewriter.Filter()
	if len(filter.Include()) <= 0 {
		ExitWithError(errors.Errorf(tr.Tr.Get("One or more files must be specified with --include")))
	}

	tracked := trackedFromExportFilter(filter)
	gitfilter := lfs.NewGitFilter(cfg)

	opts := &githistory.RewriteOptions{
		Verbose:           migrateVerbose,
		ObjectMapFilePath: objectMapFilePath,
		BlobFn: func(path string, b *gitobj.Blob) (*gitobj.Blob, error) {
			if filepath.Base(path) == ".gitattributes" {
				return b, nil
			}

			ptr, err := lfs.DecodePointer(b.Contents)
			if err != nil {
				if errors.IsNotAPointerError(err) {
					return b, nil
				}
				return nil, err
			}

			downloadPath, err := gitfilter.ObjectPath(ptr.Oid)
			if err != nil {
				return nil, err
			}

			return gitobj.NewBlobFromFile(downloadPath)
		},

		TreeCallbackFn: func(path string, t *gitobj.Tree) (*gitobj.Tree, error) {
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
			return t.Merge(&gitobj.TreeEntry{
				Name:     ".gitattributes",
				Filemode: 0100644,
				Oid:      blob,
			}), nil
		},

		UpdateRefs: true,
	}

	setupRepository()

	opts, err = rewriteOptions(args, opts, l)
	if err != nil {
		ExitWithError(err)
	}

	remote := cfg.Remote()
	if cmd.Flag("remote").Changed {
		remote = exportRemote
	}
	remoteURL := getAPIClient().Endpoints.RemoteEndpoint("download", remote).Url
	if remoteURL == "" && cmd.Flag("remote").Changed {
		ExitWithError(errors.Errorf(tr.Tr.Get("Invalid remote %s provided", remote)))
	}

	// If we have a valid remote, pre-download all objects using the Transfer Queue
	if remoteURL != "" {
		q := newDownloadQueue(getTransferManifestOperationRemote("Download", remote), remote)
		gs := lfs.NewGitScanner(cfg, func(p *lfs.WrappedPointer, err error) {
			if err != nil {
				return
			}

			if !filter.Allows(p.Name) {
				return
			}

			downloadPath, err := gitfilter.ObjectPath(p.Oid)
			if err != nil {
				return
			}

			if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
				q.Add(p.Name, downloadPath, p.Oid, p.Size, false, nil)
			}
		})
		gs.ScanRefs(opts.Include, opts.Exclude, nil)

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
		if err := performForceCheckout(l); err != nil {
			ExitWithError(err)
		}
	}

	fetchPruneCfg := lfs.NewFetchPruneConfig(cfg.Git)

	// Set our preservation time-window for objects existing on the remote to
	// 0. Because the newly rewritten commits have not yet been pushed, some
	// exported objects can still exist on the remote within the time window
	// and thus will not be pruned from the cache.
	fetchPruneCfg.FetchRecentRefsDays = 0

	// Prune our cache
	prune(fetchPruneCfg, false, false, true)
}

func performForceCheckout(l *tasklog.Logger) error {
	t := l.Waiter(fmt.Sprintf("migrate: %s", tr.Tr.Get("checkout")))
	defer t.Complete()

	return git.Checkout("", nil, true)
}

// trackedFromExportFilter returns an ordered set of strings where each entry
// is a line we intend to place in the .gitattributes file. It adds/removes the
// filter/diff/merge=lfs attributes based on patterns included/excluded in the
// given filter. Since `migrate export` removes files from Git LFS, it will
// remove attributes for included files, and add attributes for excluded files
func trackedFromExportFilter(filter *filepathfilter.Filter) *tools.OrderedSet {
	tracked := tools.NewOrderedSet()

	for _, include := range filter.Include() {
		tracked.Add(fmt.Sprintf("%s !text !filter !merge !diff", escapeAttrPattern(include)))
	}

	for _, exclude := range filter.Exclude() {
		tracked.Add(fmt.Sprintf("%s filter=lfs diff=lfs merge=lfs -text", escapeAttrPattern(exclude)))
	}

	return tracked
}
