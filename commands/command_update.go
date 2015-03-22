package commands

import (
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update local Git LFS configuration",
		Run:   updateCommand,
	}
)

// updateCommand is used for updating parts of Git LFS that reside under
// .git/lfs.
func updateCommand(cmd *cobra.Command, args []string) {
	updatePrePushHook()
}

// updatePrePushHook will force an update of the pre-push hook.
func updatePrePushHook() {
	lfs.InstallHooks(true)
	Print("Updated pre-push hook")
}

func init() {
	RootCmd.AddCommand(updateCmd)
}
