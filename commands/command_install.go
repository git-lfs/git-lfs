package commands

import (
	"os"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	forceInstall      = false
	localInstall      = false
	worktreeInstall   = false
	manualInstall     = false
	systemInstall     = false
	skipSmudgeInstall = false
	skipRepoInstall   = false
)

func installCommand(cmd *cobra.Command, args []string) {
	if err := cmdInstallOptions().Install(); err != nil {
		Print("WARNING: %s", err.Error())
		Print("Run `git lfs install --force` to reset git config.")
		os.Exit(2)
	}

	if !skipRepoInstall && (localInstall || worktreeInstall || cfg.InRepo()) {
		installHooksCommand(cmd, args)
	}

	Print("Git LFS initialized.")
}

func cmdInstallOptions() *lfs.FilterOptions {
	requireGitVersion()

	if localInstall || worktreeInstall {
		setupRepository()
	}

	switch {
	case localInstall && worktreeInstall:
		Exit("Only one of --local and --worktree options can be specified.")
	case localInstall && systemInstall:
		Exit("Only one of --local and --system options can be specified.")
	case worktreeInstall && systemInstall:
		Exit("Only one of --worktree and --system options can be specified.")
	}

	// This call will return -1 on Windows; don't warn about this there,
	// since we can't detect it correctly.
	uid := os.Geteuid()
	if systemInstall && uid != 0 && uid != -1 {
		Print("WARNING: current user is not root/admin, system install is likely to fail.")
	}

	return &lfs.FilterOptions{
		GitConfig:  cfg.GitConfig(),
		Force:      forceInstall,
		Local:      localInstall,
		Worktree:   worktreeInstall,
		System:     systemInstall,
		SkipSmudge: skipSmudgeInstall,
	}
}

func installHooksCommand(cmd *cobra.Command, args []string) {
	updateForce = forceInstall

	// TODO(@ttaylorr): this is a hack since the `git-lfs-install(1)` calls
	// into the function that implements `git-lfs-update(1)`. Given that,
	// there is no way to pass flags into that function, other than
	// hijacking the flags that `git-lfs-update(1)` already owns.
	//
	// At a later date, extract `git-lfs-update(1)`-related logic into its
	// own function, and translate this flag as a boolean argument to it.
	updateManual = manualInstall

	updateCommand(cmd, args)
}

func init() {
	RegisterCommand("install", installCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "Set the Git LFS global config, overwriting previous values.")
		cmd.Flags().BoolVarP(&localInstall, "local", "l", false, "Set the Git LFS config for the local Git repository only.")
		if git.IsGitVersionAtLeast("2.20.0") {
			cmd.Flags().BoolVarP(&worktreeInstall, "worktree", "w", false, "Set the Git LFS config for the current Git working tree, if multiple working trees are configured; otherwise, the same as --local.")
		}
		cmd.Flags().BoolVarP(&systemInstall, "system", "", false, "Set the Git LFS config in system-wide scope.")
		cmd.Flags().BoolVarP(&skipSmudgeInstall, "skip-smudge", "s", false, "Skip automatic downloading of objects on clone or pull.")
		cmd.Flags().BoolVarP(&skipRepoInstall, "skip-repo", "", false, "Skip repo setup, just install global filters.")
		cmd.Flags().BoolVarP(&manualInstall, "manual", "m", false, "Print instructions for manual install.")
		cmd.AddCommand(NewCommand("hooks", installHooksCommand))
	})
}
