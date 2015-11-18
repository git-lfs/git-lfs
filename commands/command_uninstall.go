package commands

import (
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	// uninstallCmd removes any configuration and hooks set by Git LFS.
	uninstallCmd = &cobra.Command{
		Use: "uninstall",
		Run: uninstallCommand,
	}

	// uninstallHooksCmd removes any hooks created by Git LFS.
	uninstallHooksCmd = &cobra.Command{
		Use: "hooks",
		Run: uninstallHooksCommand,
	}
)

func uninstallCommand(cmd *cobra.Command, args []string) {
	if err := lfs.UninstallFilters(); err != nil {
		Error(err.Error())
	}

	Print("Global Git LFS configuration has been removed.")

	if lfs.InRepo() {
		uninstallHooksCommand(cmd, args)
	}
}

func uninstallHooksCommand(cmd *cobra.Command, args []string) {
	if err := lfs.UninstallHooks(); err != nil {
		Error(err.Error())
	}

	Print("Hooks for this repository have been removed.")
}

func init() {
	uninstallCmd.AddCommand(uninstallHooksCmd)
	RootCmd.AddCommand(uninstallCmd)
}
