package commands

import (
	"fmt"

	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	lockCmd = &cobra.Command{
		Use: "lock",
		Run: lockCommand,
	}
)

func lockCommand(cmd *cobra.Command, args []string) {
	fmt.Println("I was run!")
}

func init() {
	RootCmd.AddCommand(lockCmd)
}
