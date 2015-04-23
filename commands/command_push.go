package commands

import (
	"github.com/spf13/cobra"
)

var (
	pushCmd = &cobra.Command{
		Use:   "push",
		Short: prePushCmd.Short,
		Run:   prePushCmd.Run,
	}
)

func init() {
	pushCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Do everything except actually send the updates")
	pushCmd.Flags().BoolVarP(&useStdin, "stdin", "s", false, "Take refs on stdin (for pre-push hook)")
	RootCmd.AddCommand(pushCmd)
}
