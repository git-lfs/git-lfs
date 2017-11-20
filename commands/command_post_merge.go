package commands

import (
	"os"

	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
)

// postMergeCommand is run through Git's post-merge hook.
//
// This hook checks that files which are lockable and not locked are made read-only,
// optimising that as best it can based on the available information.
func postMergeCommand(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		Print("This should be run through Git's post-merge hook.  Run `git lfs update` to install it.")
		os.Exit(1)
	}

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

	// The only argument this hook receives is a flag indicating whether the
	// merge was a squash merge; we don't know what files changed.
	// Whether it's squash or not is irrelevant, either way it could have
	// reset the read-only flag on files that got merged.

	tracerx.Printf("post-merge: checking write flags for all lockable files")
	// Sadly we don't get any information about what files were checked out,
	// so we have to check the entire repo
	err := lockClient.FixAllLockableFileWriteFlags()
	if err != nil {
		LoggedError(err, "Warning: post-merge locked file check failed: %v", err)
	}
}

func init() {
	RegisterCommand("post-merge", postMergeCommand, nil)
}
