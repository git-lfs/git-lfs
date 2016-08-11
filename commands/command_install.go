package commands

import (
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	forceInstall      = false
	localInstall      = false
	skipSmudgeInstall = false
)

func installCommand(cmd *cobra.Command, args []string) {
	if localInstall {
		requireInRepo()
	}

	opt := lfs.InstallOptions{Force: forceInstall, Local: localInstall}
	if skipSmudgeInstall {
		// assume the user is changing their smudge mode, so enable force implicitly
		opt.Force = true
	}

	if err := lfs.InstallFilters(opt, skipSmudgeInstall); err != nil {
		Error(err.Error())
		Exit("Run `git lfs install --force` to reset git config.")
	}

	if localInstall || lfs.InRepo() {
		installHooksCommand(cmd, args)
	}

	Print("Git LFS initialized.")
}

func installHooksCommand(cmd *cobra.Command, args []string) {
	updateForce = forceInstall
	updateCommand(cmd, args)
}

func init() {
	RegisterSubcommand(func() *cobra.Command {
		cmd := &cobra.Command{
			Use: "install",
			Run: installCommand,
		}

		cmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "Set the Git LFS global config, overwriting previous values.")
		cmd.Flags().BoolVarP(&localInstall, "local", "l", false, "Set the Git LFS config for the local Git repository only.")
		cmd.Flags().BoolVarP(&skipSmudgeInstall, "skip-smudge", "s", false, "Skip automatic downloading of objects on clone or pull.")
		cmd.AddCommand(&cobra.Command{
			Use: "hooks",
			Run: installHooksCommand,
		})
		return cmd
	})
}
