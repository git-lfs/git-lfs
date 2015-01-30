package commands

import (
	"github.com/hawser/git-hawser/hawser"
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
	if err := hawser.InstallFilters(); err != nil {
		Error(err.Error())
	}

	if hawser.InRepo() {
		initHooksCommand(cmd, args)
	}

	Print("git hawser initialized")
}

func initHooksCommand(cmd *cobra.Command, args []string) {
	if err := hawser.InstallHooks(false); err != nil {
		Error(err.Error())
	}
}

func init() {
	initCmd.AddCommand(initHooksCmd)
	RootCmd.AddCommand(initCmd)
}
