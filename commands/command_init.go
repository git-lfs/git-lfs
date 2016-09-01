package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// TODO: Remove for Git LFS v2.0 https://github.com/github/git-lfs/issues/839
func initCommand(cmd *cobra.Command, args []string) {
	fmt.Fprintf(os.Stderr, "WARNING: 'git lfs init' is deprecated. Use 'git lfs install' now.\n")
	installCommand(cmd, args)
}

func initHooksCommand(cmd *cobra.Command, args []string) {
	fmt.Fprintf(os.Stderr, "WARNING: 'git lfs init' is deprecated. Use 'git lfs install' now.\n")
	installHooksCommand(cmd, args)
}

func init() {
	RegisterCommand("init", initCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "Set the Git LFS global config, overwriting previous values.")
		cmd.Flags().BoolVarP(&localInstall, "local", "l", false, "Set the Git LFS config for the local Git repository only.")
		cmd.Flags().BoolVarP(&skipSmudgeInstall, "skip-smudge", "s", false, "Skip automatic downloading of objects on clone or pull.")
		cmd.AddCommand(NewCommand("hooks", initHooksCommand))
	})
}
