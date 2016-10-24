package commands

import (
	"os"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/locking"
	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
)

// postCommitCommand is run through Git's post-commit hook. The hook passes
// no arguments.
// This hook checks that files which are lockable and not locked are made read-only,
// optimising that based on what was added / modified in the commit.
func postCommitCommand(cmd *cobra.Command, args []string) {
	requireGitVersion()

	// Skip this hook if no lockable patterns have been configured
	if len(locking.GetLockablePatterns()) == 0 {
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
	err = locking.FixLockableFileWriteFlags(files)
	if err != nil {
		LoggedError(err, "Warning: post-commit locked file check failed: %v", err)
	}

}

func init() {
	RegisterCommand("post-commit", postCommitCommand, nil)
}
