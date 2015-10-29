package commands

import (
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	// uninitCmd removes any configuration and hooks set by Git LFS.
	uninitCmd = &cobra.Command{
		Use: "uninit",
		Run: uninitCommand,
	}

	// uninitHooksCmd removes any hooks created by Git LFS.
	uninitHooksCmd = &cobra.Command{
		Use: "hooks",
		Run: uninitHooksCommand,
	}
)

func uninitCommand(cmd *cobra.Command, args []string) {
	if err := lfs.UninstallFilters(); err != nil {
		Error(err.Error())
	}

	Print("Global Git LFS configuration has been removed.")

	if lfs.InRepo() {
		uninitHooksCommand(cmd, args)
	}
}

func uninitHooksCommand(cmd *cobra.Command, args []string) {
	if err := lfs.UninstallHooks(); err != nil {
		Error(err.Error())
	}

	Print("Hooks for this repository have been removed.")
}

func init() {
	uninitCmd.AddCommand(uninitHooksCmd)
	RootCmd.AddCommand(uninitCmd)
}
