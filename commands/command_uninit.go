package commands

import (
	"fmt"
	"os"

	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

// TODO: Remove for Git LFS v2.0

var (
	// uninitCmd removes any configuration and hooks set by Git LFS.
	uninitCmd = &cobra.Command{
		Use: "uninit",
		Run: uninitCommand,
	}

	// uninitHooksCmd removes any hooks created by Git LFS.
	uninitHooksCmd = &cobra.Command{
		Use: "hooks",
		Run: uninitHooksCommand,
	}
)

func uninitCommand(cmd *cobra.Command, args []string) {
	fmt.Fprintf(os.Stderr, "WARNING: 'git lfs uninit' is deprecated. Use 'git lfs uninstall' now.\n")
	uninstallCommand(cmd, args)
}

func uninitHooksCommand(cmd *cobra.Command, args []string) {
	fmt.Fprintf(os.Stderr, "WARNING: 'git lfs uninit' is deprecated. Use 'git lfs uninstall' now.\n")
	uninstallHooksCommand(cmd, args)
}

func init() {
	uninitCmd.AddCommand(uninitHooksCmd)
	RootCmd.AddCommand(uninitCmd)
}
