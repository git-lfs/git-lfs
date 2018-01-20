package commands

import (
	"os"

	"github.com/git-lfs/git-lfs/git"
	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
)

// postCommitCommand is run through Git's post-commit hook. The hook passes
// no arguments.
// This hook checks that files which are lockable and not locked are made read-only,
// optimising that based on what was added / modified in the commit.
// This is mainly to catch added files, since modified files should already be
// locked. If we didn't do this, any added files would remain read/write on disk
// even without a lock unless something else checked.
func postCommitCommand(cmd *cobra.Command, args []string) {

	// Skip entire hook if lockable read only feature is disabled
	if !cfg.SetLockableFilesReadOnly() {
		os.Exit(0)
	}

	requireGitVersion()

	lockClient := newLockClient()

	// Skip this hook if no lockable patterns have been configured
	if len(lockClient.GetLockablePatterns()) == 0 {
		os.Exit(0)
	}

	tracerx.Printf("post-commit: checking file write flags at HEAD")
	// We can speed things up by looking at what changed in
	// HEAD, and only checking those lockable files
	files, err := git.GetFilesChanged("HEAD", "")

	if err != nil {
		LoggedError(err, "Warning: post-commit failed: %v", err)
		os.Exit(1)
	}
	tracerx.Printf("post-commit: checking write flags on %v", files)
	err = lockClient.FixLockableFileWriteFlags(files)
	if err != nil {
		LoggedError(err, "Warning: post-commit locked file check failed: %v", err)
	}

}

func init() {
	RegisterCommand("post-commit", postCommitCommand, nil)
}
