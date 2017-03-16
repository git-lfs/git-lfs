package commands

import (
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

	entries, errs, err := scanIndex(scanIndexAt)
	if err != nil {
		ExitWithError(err)
	}

	staged := make([]*lfs.DiffIndexEntry, 0)
	unstaged := make([]*lfs.DiffIndexEntry, 0)

L:
	for {
		select {
		case entry, ok := <-entries:
			if !ok {
				break L
			}

			switch entry.Status {
			case lfs.StatusModification:
				unstaged = append(unstaged, entry)
			default:
				staged = append(staged, entry)
			}
		case err := <-errs:
			if err != nil {
				ExitWithError(err)
			}
		}
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

func scanIndex(ref string) (<-chan *lfs.DiffIndexEntry, <-chan error, error) {
	uncached, err := lfs.NewDiffIndexScanner(ref, false)
	if err != nil {
		return nil, nil, err
	}

	cached, err := lfs.NewDiffIndexScanner(ref, true)
	if err != nil {
		return nil, nil, err
	}

	entries := make(chan *lfs.DiffIndexEntry)
	errs := make(chan error)

	go func() {
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
					entries <- entry
					seenNames[name] = struct{}{}
				}
			}

			if err := scanner.Err(); err != nil {
				errs <- err
			}
		}

		close(entries)
		close(errs)
	}()

	return entries, errs, nil
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
	gitscanner := lfs.NewGitScanner(func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			ExitWithError(err)
		}

		switch p.Status {
		case "R", "C":
			Print("%s  %s -> %s", p.Status, p.SrcName, p.Name)
		case "M":
			Print(" %s %s", p.Status, p.Name)
		default:
			Print("%s  %s", p.Status, p.Name)
		}
	})
	defer gitscanner.Close()

	if err := gitscanner.ScanIndex(ref, nil); err != nil {
		ExitWithError(err)
	}
}

func init() {
	RegisterCommand("status", statusCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&porcelain, "porcelain", "p", false, "Give the output in an easy-to-parse format for scripts.")
	})
}
