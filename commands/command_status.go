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

	scanIndexAt := "HEAD"
	if ref == nil {
		scanIndexAt = git.RefBeforeFirstCommit
	}

	if porcelain {
		porcelainStagedPointers(scanIndexAt)
		return
	}

	statusScanRefRange(ref)

	staged, unstaged, err := scanIndex(scanIndexAt)
	if err != nil {
		ExitWithError(err)
	}

	Print("\nGit LFS objects to be committed:\n")
	for _, entry := range staged {
		switch entry.Status {
		case lfs.StatusRename, lfs.StatusCopy:
			Print("\t%s -> %s", entry.SrcName, entry.DstName)
		default:
			Print("\t%s", entry.SrcName)
		}
	}

	Print("\nGit LFS objects not staged for commit:\n")
	for _, entry := range unstaged {
		Print("\t%s", entry.SrcName)
	}

	Print("")
}

func scanIndex(ref string) (staged, unstaged []*lfs.DiffIndexEntry, err error) {
	uncached, err := lfs.NewDiffIndexScanner(ref, false)
	if err != nil {
		return nil, nil, err
	}

	cached, err := lfs.NewDiffIndexScanner(ref, true)
	if err != nil {
		return nil, nil, err
	}

	seenNames := make(map[string]struct{}, 0)

	for _, scanner := range []*lfs.DiffIndexScanner{
		uncached, cached,
	} {
		for scanner.Scan() {
			entry := scanner.Entry()

			name := entry.DstName
			if len(name) == 0 {
				name = entry.SrcName
			}

			if _, seen := seenNames[name]; !seen {
				switch entry.Status {
				case lfs.StatusModification:
					unstaged = append(unstaged, entry)
				default:
					staged = append(staged, entry)
				}

				seenNames[name] = struct{}{}
			}
		}

		if err := scanner.Err(); err != nil {
			return nil, nil, err
		}
	}
	return
}

func statusScanRefRange(ref *git.Ref) {
	if ref == nil {
		return
	}

	Print("On branch %s", ref.Name)

	remoteRef, err := git.CurrentRemoteRef()
	if err != nil {
		return
	}

	gitscanner := lfs.NewGitScanner(func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			Panic(err, "Could not scan for Git LFS objects")
			return
		}

		Print("\t%s (%s)", p.Name)
	})
	defer gitscanner.Close()

	Print("Git LFS objects to be pushed to %s:\n", remoteRef.Name)
	if err := gitscanner.ScanRefRange(ref.Sha, "^"+remoteRef.Sha, nil); err != nil {
		Panic(err, "Could not scan for Git LFS objects")
	}

}

func porcelainStagedPointers(ref string) {
	staged, unstaged, err := scanIndex(ref)
	if err != nil {
		ExitWithError(err)
	}

	for _, entry := range append(unstaged, staged...) {
		Print(porcelainStatusLine(entry))
	}
}

func porcelainStatusLine(entry *lfs.DiffIndexEntry) string {
	switch entry.Status {
	case lfs.StatusRename, lfs.StatusCopy:
		return fmt.Sprintf("%s  %s -> %s", entry.Status, entry.SrcName, entry.DstName)
	case lfs.StatusModification:
		return fmt.Sprintf(" %s %s", entry.Status, entry.SrcName)
	}

	return fmt.Sprintf("%s  %s", entry.Status, entry.SrcName)
}

func init() {
	RegisterCommand("status", statusCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&porcelain, "porcelain", "p", false, "Give the output in an easy-to-parse format for scripts.")
	})
}
