package commands

import (
	"github.com/github/git-lfs/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/github/git-lfs/lfs"
)

var (
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize the default Git LFS configuration",
		Run:   initCommand,
	}

	initHooksCmd = &cobra.Command{
		Use:   "hooks",
		Short: "Initialize hooks for the current repository",
		Run:   initHooksCommand,
	}
)

func initCommand(cmd *cobra.Command, args []string) {
	if err := lfs.InstallFilters(); err != nil {
		Error(err.Error())
	}

	if lfs.InRepo() {
		initHooksCommand(cmd, args)
	}

	Print("git lfs initialized")
}

func initHooksCommand(cmd *cobra.Command, args []string) {
	if err := lfs.InstallHooks(false); err != nil {
		Error(err.Error())
	}
}

func init() {
	initCmd.AddCommand(initHooksCmd)
	RootCmd.AddCommand(initCmd)
}
