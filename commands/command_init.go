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
)

func initCommand(cmd *cobra.Command, args []string) {
	var sub string
	if len(args) > 0 {
		sub = args[0]
	}

	switch sub {
	case "hooks":
		hookInit()
	default:
		runInit()
	}

	Print("git media initialized")
}

func runInit() {
	globalInit()

	if gitmedia.InRepo() {
		hookInit()
	}
}

func globalInit() {
	if err := gitmedia.InstallFilters(); err != nil {
		Error(err.Error())
	}
}

func hookInit() {
	if err := gitmedia.InstallHooks(); err != nil {
		Error(err.Error())
	}
}

func init() {
	RootCmd.AddCommand(initCmd)
}
