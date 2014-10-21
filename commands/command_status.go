package commands

import (
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/scanner"
	"github.com/spf13/cobra"
)

var (
	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show information about git media files that would be pushed",
		Run:   statusCommand,
	}
)

func statusCommand(cmd *cobra.Command, args []string) {
	ref, err := gitmedia.CurrentRef()
	if err != nil {
		Panic(err, "Could not ls-files")
	}

	pointers, err := scanner.Scan(ref, "^origin/HEAD")
	if err != nil {
		Panic(err, "Could not scan for git media files")
	}

	for _, p := range pointers {
		Print(p.Name)
	}
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
