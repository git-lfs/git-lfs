package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// TODO: Remove for Git LFS v2.0 https://github.com/github/git-lfs/issues/839

// uninitCmd removes any configuration and hooks set by Git LFS.
func uninitCommand(cmd *cobra.Command, args []string) {
	fmt.Fprintf(os.Stderr, "WARNING: 'git lfs uninit' is deprecated. Use 'git lfs uninstall' now.\n")
	uninstallCommand(cmd, args)
}

// uninitHooksCmd removes any hooks created by Git LFS.
func uninitHooksCommand(cmd *cobra.Command, args []string) {
	fmt.Fprintf(os.Stderr, "WARNING: 'git lfs uninit' is deprecated. Use 'git lfs uninstall' now.\n")
	uninstallHooksCommand(cmd, args)
}

func init() {
	RegisterCommand("uninit", uninitCommand, func(cmd *cobra.Command) {
		cmd.AddCommand(NewCommand("hooks", uninitHooksCommand))
	})
}
