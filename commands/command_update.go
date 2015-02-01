package commands

import (
	"github.com/hawser/git-hawser/hawser"
	"github.com/spf13/cobra"
)

var (
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update local hawser configuration",
		Run:   updateCommand,
	}
)

// updateCommand is used for updating parts of hawser that reside
// under .git/hawser.
func updateCommand(cmd *cobra.Command, args []string) {
	updatePrePushHook()
}

// updatePrePushHook will force an update of the pre-push hook.
func updatePrePushHook() {
	hawser.InstallHooks(true)
	Print("Updated pre-push hook")
}

func init() {
	RootCmd.AddCommand(updateCmd)
}
