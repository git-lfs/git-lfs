package commands

import (
	"github.com/github/git-media/git"
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

	stagedPointers, err := scanner.ScanIndex()
	if err != nil {
		Panic(err, "Could not scan staging for git media files")
	}

	branch, err := git.CurrentBranch()
	if err != nil {
		Panic(err, "Could not get current branch")
	}
	Print("On branch %s", branch)

	remote, err := git.CurrentRemote()
	if err != nil {
		Panic(err, "Could not get current remote branch")
	}

	Print("Media file changes to be pushed to %s:\n", remote)
	for _, p := range pointers {
		Print("\t%s (%d bytes)", p.Name, p.Size)
	}

	Print("\nMedia file changes to be committed:\n")
	for _, p := range stagedPointers {
		if p.Status != "M" {
			Print("\t%s (%d bytes)", p.Name, p.Size)
		}
	}

	Print("\nMedia file changes not staged for commit:\n")
	for _, p := range stagedPointers {
		if p.Status == "M" {
			Print("\t%s", p.Name)
		}
	}

	Print("")
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
