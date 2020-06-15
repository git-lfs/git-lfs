package commands

import (
	"github.com/git-lfs/git-lfs/git"
	"github.com/spf13/cobra"
)

// uninstallCmd removes any configuration and hooks set by Git LFS.
func uninstallCommand(cmd *cobra.Command, args []string) {
	if err := cmdInstallOptions().Uninstall(); err != nil {
		Print("WARNING: %s", err.Error())
	}

	if !skipRepoInstall && (localInstall || worktreeInstall || cfg.InRepo()) {
		uninstallHooksCommand(cmd, args)
	}

	if systemInstall {
		Print("System Git LFS configuration has been removed.")
	} else if !(localInstall || worktreeInstall) {
		Print("Global Git LFS configuration has been removed.")
	}
}

// uninstallHooksCmd removes any hooks created by Git LFS.
func uninstallHooksCommand(cmd *cobra.Command, args []string) {
	if err := uninstallHooks(); err != nil {
		Error(err.Error())
	}

	Print("Hooks for this repository have been removed.")
}

func init() {
	RegisterCommand("uninstall", uninstallCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&localInstall, "local", "l", false, "Remove the Git LFS config for the local Git repository only.")
		if git.IsGitVersionAtLeast("2.20.0") {
			cmd.Flags().BoolVarP(&worktreeInstall, "worktree", "w", false, "Remove the Git LFS config for the current Git working tree, if multiple working trees are configured; otherwise, the same as --local.")
		}
		cmd.Flags().BoolVarP(&systemInstall, "system", "", false, "Remove the Git LFS config in system-wide scope.")
		cmd.Flags().BoolVarP(&skipRepoInstall, "skip-repo", "", false, "Skip repo setup, just uninstall global filters.")
		cmd.AddCommand(NewCommand("hooks", uninstallHooksCommand))
	})
}
