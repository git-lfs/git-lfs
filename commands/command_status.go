package commands

import (
	"fmt"
	"github.com/github/git-media/git"
	"github.com/github/git-media/scanner"
	"github.com/spf13/cobra"
)

var (
	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show information about git media files that would be pushed",
		Run:   statusCommand,
	}
	porcelain = false
)

func statusCommand(cmd *cobra.Command, args []string) {
	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not get the current ref")
	}

	stagedPointers, err := scanner.ScanIndex()
	if err != nil {
		Panic(err, "Could not scan staging for git media files")
	}

	if porcelain {
		for _, p := range stagedPointers {
			switch p.Status {
			case "R", "C":
				Print("%s  %s -> %s %d", p.Status, p.SrcName, p.Name, p.Size)
			case "M":
				Print(" %s %s %d", p.Status, p.Name, p.Size)
			default:
				Print("%s  %s %d", p.Status, p.Name, p.Size)
			}
		}
		return
	}

	branch, err := git.CurrentBranch()
	if err != nil {
		Panic(err, "Could not get current branch")
	}
	Print("On branch %s", branch)

	remoteRef, err := git.CurrentRemoteRef()
	if err == nil {

		pointers, err := scanner.Scan(ref, "^"+remoteRef)
		if err != nil {
			Panic(err, "Could not scan for git media files")
		}

		remote, err := git.CurrentRemote()
		if err != nil {
			Panic(err, "Could not get current remote branch")
		}

		Print("Media file changes to be pushed to %s:\n", remote)
		for _, p := range pointers {
			Print("\t%s (%s)", p.Name, humanizeBytes(p.Size))
		}
	}

	Print("\nMedia file changes to be committed:\n")
	for _, p := range stagedPointers {
		switch p.Status {
		case "R", "C":
			Print("\t%s -> %s (%s)", p.SrcName, p.Name, humanizeBytes(p.Size))
		case "M":
		default:
			Print("\t%s (%s)", p.Name, humanizeBytes(p.Size))
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

var byteUnits = []string{"B", "KB", "MB", "GB", "TB"}

func humanizeBytes(bytes int64) string {
	var output string
	size := float64(bytes)

	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}

	for _, unit := range byteUnits {
		if size < 1024.0 {
			output = fmt.Sprintf("%3.1f %s", size, unit)
			break
		}
		size /= 1024.0
	}
	return output
}

func init() {
	statusCmd.Flags().BoolVarP(&porcelain, "porcelain", "p", false, "Give the output in an easy-to-parse format for scripts.")
	RootCmd.AddCommand(statusCmd)
}
