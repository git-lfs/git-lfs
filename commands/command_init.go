package commands

import (
	"fmt"
	"os"

	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

// TODO: Remove for Git LFS v2.0

var (
	initCmd = &cobra.Command{
		Use: "init",
		Run: initCommand,
	}

	initHooksCmd = &cobra.Command{
		Use: "hooks",
		Run: initHooksCommand,
	}
)

func initCommand(cmd *cobra.Command, args []string) {
	fmt.Fprintf(os.Stderr, "WARNING: 'git lfs init' is deprecated. Use 'git lfs install' now.\n")
	installCommand(cmd, args)
}

func initHooksCommand(cmd *cobra.Command, args []string) {
	fmt.Fprintf(os.Stderr, "WARNING: 'git lfs init' is deprecated. Use 'git lfs install' now.\n")
	installHooksCommand(cmd, args)
}

func init() {
	initCmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "Set the Git LFS global config, overwriting previous values.")
	initCmd.Flags().BoolVarP(&localInstall, "local", "l", false, "Set the Git LFS config for the local Git repository only.")
	initCmd.Flags().BoolVarP(&skipSmudgeInstall, "skip-smudge", "s", false, "Skip automatic downloading of objects on clone or pull.")
	initCmd.AddCommand(initHooksCmd)
	RootCmd.AddCommand(initCmd)
}
