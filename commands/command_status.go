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
		Panic(err, "Could not calculate status")
	}

	remoteRef, err := gitmedia.CurrentRemoteRef()
	if err != nil {
		Panic(err, "Could not calculate status")
	}

	pointers, err := scanner.Scan(ref, "^"+remoteRef)
	if err != nil {
		Panic(err, "Could not scan for git media files")
	}

	for _, p := range pointers {
		Print("%s %d", p.Name, p.Size)
	}

	pointers, err = scanner.ScanStaging()
	if err != nil {
		Panic(err, "Could not scan staging for git media files")
	}

	for _, p := range pointers {
		Print("%s %d", p.Name, p.Size)
	}
}

// hi

func init() {
	RootCmd.AddCommand(statusCmd)
}
