package commands

import "github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"

// TODO: Remove for Git LFS v2.0 https://github.com/github/git-lfs/issues/839

var (
	// uninitCmd removes any configuration and hooks set by Git LFS.
	uninitCmd = &cobra.Command{
		Use:        "uninit",
		Deprecated: "Use 'git lfs uninstall' now",
		Run:        uninitCommand,
	}

	// uninitHooksCmd removes any hooks created by Git LFS.
	uninitHooksCmd = &cobra.Command{
		Use:        "hooks",
		Deprecated: "Use 'git lfs uninstall' now",
		Run:        uninitHooksCommand,
	}
)

func uninitCommand(cmd *cobra.Command, args []string) {
	uninstallCommand(cmd, args)
}

func uninitHooksCommand(cmd *cobra.Command, args []string) {
	uninstallHooksCommand(cmd, args)
}

func init() {
	uninitCmd.AddCommand(uninitHooksCmd)
	RootCmd.AddCommand(uninitCmd)
}
