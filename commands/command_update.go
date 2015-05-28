package commands

import (
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update local Git LFS configuration",
		Run:   updateCommand,
	}

	updateForce = false
)

// updateCommand is used for updating parts of Git LFS that reside under
// .git/lfs.
func updateCommand(cmd *cobra.Command, args []string) {
	updatePrePushHook()
}

// updatePrePushHook will force an update of the pre-push hook.
func updatePrePushHook() {
	if err := lfs.InstallHooks(updateForce); err != nil {
		Error(err.Error())
		Print("Run `git lfs update --force` to overwrite this hook.")
	} else {
		Print("Updated pre-push hook")
	}

}

func init() {
	updateCmd.Flags().BoolVarP(&updateForce, "force", "f", false, "Overwrite hooks.")
	RootCmd.AddCommand(updateCmd)
}
