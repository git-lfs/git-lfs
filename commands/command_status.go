package commands

import (
	"fmt"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	porcelain = false
)

func statusCommand(cmd *cobra.Command, args []string) {
	requireInRepo()

	// tolerate errors getting ref so this works before first commit
	ref, _ := git.CurrentRef()

	gitscanner := lfs.NewGitScanner()
	defer gitscanner.Close()

	scanIndexAt := "HEAD"
	if ref == nil {
		scanIndexAt = git.RefBeforeFirstCommit
	}

	stagedPointers, err := gitscanner.ScanIndex(scanIndexAt)
	if err != nil {
		Panic(err, "Could not scan staging for Git LFS objects")
	}

	if porcelain {
		for p := range stagedPointers.Results {
			switch p.Status {
			case "R", "C":
				Print("%s  %s -> %s %d", p.Status, p.SrcName, p.Name, p.Size)
			case "M":
				Print(" %s %s %d", p.Status, p.Name, p.Size)
			default:
				Print("%s  %s %d", p.Status, p.Name, p.Size)
			}
		}

		if err := stagedPointers.Wait(); err != nil {
			ExitWithError(err)
		}
		return
	}

	if ref != nil {
		Print("On branch %s", ref.Name)

		remoteRef, err := git.CurrentRemoteRef()
		if err == nil {
			pointerCh, err := gitscanner.ScanRefRange(ref.Sha, "^"+remoteRef.Sha)
			if err != nil {
				Panic(err, "Could not scan for Git LFS objects")
			}

			Print("Git LFS objects to be pushed to %s:\n", remoteRef.Name)
			for p := range pointerCh.Results {
				Print("\t%s (%s)", p.Name, humanizeBytes(p.Size))
			}

			if err := pointerCh.Wait(); err != nil {
				Panic(err, "Could not scan for Git LFS objects")
			}
		}
	}

	Print("\nGit LFS objects to be committed:\n")
	var unstagedPointers []*lfs.WrappedPointer
	for p := range stagedPointers.Results {
		switch p.Status {
		case "R", "C":
			Print("\t%s -> %s (%s)", p.SrcName, p.Name, humanizeBytes(p.Size))
		case "M":
			unstagedPointers = append(unstagedPointers, p)
		default:
			Print("\t%s (%s)", p.Name, humanizeBytes(p.Size))
		}
	}

	Print("\nGit LFS objects not staged for commit:\n")
	for _, p := range unstagedPointers {
		if p.Status == "M" {
			Print("\t%s", p.Name)
		}
	}

	if err := stagedPointers.Wait(); err != nil {
		ExitWithError(err)
	} else {
		Print("")
	}
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
