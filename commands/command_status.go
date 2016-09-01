package commands

import (
	"fmt"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	porcelain = false
)

func statusCommand(cmd *cobra.Command, args []string) {
	requireInRepo()

	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not get the current ref")
	}

	stagedPointers, err := lfs.ScanIndex()
	if err != nil {
		Panic(err, "Could not scan staging for Git LFS objects")
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

	Print("On branch %s", ref.Name)

	remoteRef, err := git.CurrentRemoteRef()
	if err == nil {

		pointers, err := lfs.ScanRefs(ref.Sha, "^"+remoteRef.Sha, nil)
		if err != nil {
			Panic(err, "Could not scan for Git LFS objects")
		}

		Print("Git LFS objects to be pushed to %s:\n", remoteRef.Name)
		for _, p := range pointers {
			Print("\t%s (%s)", p.Name, humanizeBytes(p.Size))
		}
	}

	Print("\nGit LFS objects to be committed:\n")
	for _, p := range stagedPointers {
		switch p.Status {
		case "R", "C":
			Print("\t%s -> %s (%s)", p.SrcName, p.Name, humanizeBytes(p.Size))
		case "M":
		default:
			Print("\t%s (%s)", p.Name, humanizeBytes(p.Size))
		}
	}

	Print("\nGit LFS objects not staged for commit:\n")
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
	RegisterCommand("status", statusCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&porcelain, "porcelain", "p", false, "Give the output in an easy-to-parse format for scripts.")
	})
}
