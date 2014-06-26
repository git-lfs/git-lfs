package commands

import (
	"github.com/github/git-media/gitmedia"
	"github.com/spf13/cobra"
)

var (
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize the default Git Media configuration",
		Run:   initCommand,
	}

	initHooksCmd = &cobra.Command{
		Use:   "hooks",
		Short: "Initialize hooks for the current repository",
		Run:   initHooksCommand,
	}
)

func initCommand(cmd *cobra.Command, args []string) {
	if err := gitmedia.InstallFilters(); err != nil {
		Error(err.Error())
	}

	if gitmedia.InRepo() {
		initHooksCommand(cmd, args)
	}

	Print("git media initialized")
}

func initHooksCommand(cmd *cobra.Command, args []string) {
	if err := gitmedia.InstallHooks(); err != nil {
		Error(err.Error())
	}
}

func init() {
	initCmd.AddCommand(initHooksCmd)
	RootCmd.AddCommand(initCmd)
}
