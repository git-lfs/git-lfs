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

	forceInit      = false
	localInit      = false
	skipSmudgeInit = false
)

func initCommand(cmd *cobra.Command, args []string) {
	if localInit {
		requireInRepo()
	}

	opt := lfs.InstallOptions{Force: forceInit, Local: localInit}
	if skipSmudgeInit {
		// assume the user is changing their smudge mode, so enable force implicitly
		opt.Force = true
	}

	if err := lfs.InstallFilters(opt, skipSmudgeInit); err != nil {
		Error(err.Error())
		Exit("Run `git lfs init --force` to reset git config.")
	}

	if localInit || lfs.InRepo() {
		initHooksCommand(cmd, args)
	}

	Print("Git LFS initialized.")
}

func initHooksCommand(cmd *cobra.Command, args []string) {
	updateForce = forceInit
	updateCommand(cmd, args)
}

func init() {
	initCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Set the Git LFS global config, overwriting previous values.")
	initCmd.Flags().BoolVarP(&localInit, "local", "l", false, "Set the Git LFS config for the local Git repository only.")
	initCmd.Flags().BoolVarP(&skipSmudgeInit, "skip-smudge", "s", false, "Skip automatic downloading of objects on clone or pull.")
	initCmd.AddCommand(initHooksCmd)
	RootCmd.AddCommand(initCmd)
}
