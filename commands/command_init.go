package commands

import (
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	initCmd = &cobra.Command{
		Use: "init",
		Run: initCommand,
	}

	initHooksCmd = &cobra.Command{
		Use: "hooks",
		Run: initHooksCommand,
	}

	forceInit = false
)

func initCommand(cmd *cobra.Command, args []string) {
	if err := lfs.InstallFilters(forceInit); err != nil {
		Error(err.Error())
		Exit("Run `git lfs init --force` to reset git config.")
	}

	initHooksCommand(cmd, args)
	Print("Git LFS initialized.")
}

func initHooksCommand(cmd *cobra.Command, args []string) {
	updateForce = forceInit
	updateCommand(cmd, args)
}

func init() {
	initCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Set the Git LFS global config, overwriting previous values.")
	initCmd.AddCommand(initHooksCmd)
	RootCmd.AddCommand(initCmd)
}
