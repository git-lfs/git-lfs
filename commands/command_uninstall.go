package commands

import (
	"github.com/spf13/cobra"
)

// uninstallCmd removes any configuration and hooks set by Git LFS.
func uninstallCommand(cmd *cobra.Command, args []string) {
	if err := cmdInstallOptions().Uninstall(); err != nil {
		Error(err.Error())
	}

	if localInstall || cfg.InRepo() {
		uninstallHooksCommand(cmd, args)
	}

	if !localInstall {
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
		cmd.Flags().BoolVarP(&localInstall, "local", "l", false, "Set the Git LFS config for the local Git repository only.")
		cmd.Flags().BoolVarP(&systemInstall, "system", "", false, "Set the Git LFS config in system-wide scope.")
		cmd.AddCommand(NewCommand("hooks", uninstallHooksCommand))
	})
}
