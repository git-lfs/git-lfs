package commands

import (
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
	"os"
)

var (
	installCmd = &cobra.Command{
		Use: "install",
		Run: installCommand,
	}

	installHooksCmd = &cobra.Command{
		Use: "hooks",
		Run: installHooksCommand,
	}

	forceInstall      = false
	localInstall      = false
	systemInstall     = false
	skipSmudgeInstall = false
)

func installCommand(cmd *cobra.Command, args []string) {
	if localInstall {
		requireInRepo()
	}

	if systemInstall && os.Geteuid() != 0 {
		Print("WARNING: current user is not root/admin, system install is likely to fail.")
	}

	if localInstall && systemInstall {
		Exit("Only one of --local and --system options can be specified.")
	}

	opt := lfs.InstallOptions{Force: forceInstall, Local: localInstall, System: systemInstall}
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
	installCmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "Set the Git LFS global config, overwriting previous values.")
	installCmd.Flags().BoolVarP(&localInstall, "local", "l", false, "Set the Git LFS config for the local Git repository only.")
	// TODO - -s was already taken by --skip-smudge, what should --system be shortened-to?
	installCmd.Flags().BoolVarP(&systemInstall, "system", "", false, "Set the Git LFS config in system-wide scope.")
	installCmd.Flags().BoolVarP(&skipSmudgeInstall, "skip-smudge", "s", false, "Skip automatic downloading of objects on clone or pull.")
	installCmd.AddCommand(installHooksCmd)
	RootCmd.AddCommand(installCmd)
}
