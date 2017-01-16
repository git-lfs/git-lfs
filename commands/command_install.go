package commands

import (
	"os"

	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/localstorage"
	"github.com/spf13/cobra"
)

var (
	forceInstall      = false
	localInstall      = false
	systemInstall     = false
	skipSmudgeInstall = false
	skipRepoInstall   = false
)

func installCommand(cmd *cobra.Command, args []string) {
	requireGitVersion()

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

	if !skipRepoInstall && (localInstall || lfs.InRepo()) {
		localstorage.InitStorageOrFail()
		installHooksCommand(cmd, args)
	}

	Print("Git LFS initialized.")
}

func installHooksCommand(cmd *cobra.Command, args []string) {
	updateForce = forceInstall
	updateCommand(cmd, args)
}

func init() {
	RegisterCommand("install", installCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "Set the Git LFS global config, overwriting previous values.")
		cmd.Flags().BoolVarP(&localInstall, "local", "l", false, "Set the Git LFS config for the local Git repository only.")
		cmd.Flags().BoolVarP(&systemInstall, "system", "", false, "Set the Git LFS config in system-wide scope.")
		cmd.Flags().BoolVarP(&skipSmudgeInstall, "skip-smudge", "s", false, "Skip automatic downloading of objects on clone or pull.")
		cmd.Flags().BoolVarP(&skipRepoInstall, "skip-repo", "", false, "Skip repo setup, just install global filters.")
		cmd.AddCommand(NewCommand("hooks", installHooksCommand))
		cmd.PreRun = setupLocalStorage
	})
}
