package commands

import (
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/spf13/cobra"
)

var (
	dedupFlags = struct {
		test bool
	}{}
	dedupStats = &struct {
		totalProcessedCount int64
		totalProcessedSize  int64
	}{}
)

func dedupTestCommand(*cobra.Command, []string) {
	setupRepository()

	if supported, err := tools.CheckCloneFileSupported(cfg.TempDir()); err != nil || !supported {
		if err == nil {
			err = errors.New("Unknown reason.")
		}
		Exit("This system does not support deduplication. %s", err)
	}

	if len(cfg.Extensions()) > 0 {
		Exit("This platform supports file de-duplication, however, Git LFS extensions are configured and therefore de-duplication can not be used.")
	}

	Print("OK: This platform and repository support file de-duplication.")
}

func dedupCommand(cmd *cobra.Command, args []string) {
	if dedupFlags.test {
		dedupTestCommand(cmd, args)
		return
	}

	setupRepository()
	if gitDir, err := git.GitDir(); err != nil {
		ExitWithError(err)
	} else if supported, err := tools.CheckCloneFileSupported(gitDir); err != nil || !supported {
		Exit("This system does not support deduplication.")
	}

	if len(cfg.Extensions()) > 0 {
		Exit("This platform supports file de-duplication, however, Git LFS extensions are configured and therefore de-duplication can not be used.")
	}

	if dirty, err := git.IsWorkingCopyDirty(); err != nil {
		ExitWithError(err)
	} else if dirty {
		Exit("Working tree is dirty. Please commit or reset your change.")
	}

	// We assume working tree is clean.
	gitScanner := lfs.NewGitScanner(config.New(), func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			Exit("Could not scan for Git LFS tree: %s", err)
			return
		}

		if success, err := dedup(p); err != nil {
			Error("Skipped: %s (Size: %d)\n          %s", p.Name, p.Size, err)
		} else if !success {
			Error("Skipped: %s (Size: %d)", p.Name, p.Size)
		} else if success {
			Print("Success: %s (Size: %d)", p.Name, p.Size)

			atomic.AddInt64(&dedupStats.totalProcessedCount, 1)
			atomic.AddInt64(&dedupStats.totalProcessedSize, p.Size)
		}
	})
	defer gitScanner.Close()

	if err := gitScanner.ScanTree("HEAD"); err != nil {
		ExitWithError(err)
	}

	Print("\n\nSuccessfully finished.\n"+
		"  De-duplicated  size: %d bytes\n"+
		"                count: %d",
		dedupStats.totalProcessedSize,
		dedupStats.totalProcessedCount)
}

// dedup executes
// Precondition: working tree MUST clean. We can replace working tree files from mediafile safely.
func dedup(p *lfs.WrappedPointer) (success bool, err error) {
	// PRECONDITION, check ofs object exists or skip this file.
	if !cfg.LFSObjectExists(p.Oid, p.Size) { // Not exists,
		// Basically, this is not happens because executing 'git status' in `git.IsWorkingCopyDirty()` recover it.
		return false, errors.New("mediafile is not exist")
	}

	// DO de-dup
	// Gather original state
	originalStat, err := os.Stat(p.Name)
	if err != nil {
		return false, err
	}

	// Do clone
	srcFile := cfg.Filesystem().ObjectPathname(p.Oid)
	dstFile := filepath.Join(cfg.LocalWorkingDir(), p.Name)

	// Clone the file. This overwrites the destination if it exists.
	if ok, err := tools.CloneFileByPath(dstFile, srcFile); err != nil {
		return false, err
	} else if !ok {
		return false, errors.Errorf("unknown clone file error")
	}

	// Recover original state
	if err := os.Chmod(dstFile, originalStat.Mode()); err != nil {
		return false, err
	}

	return true, nil
}

func init() {
	RegisterCommand("dedup", dedupCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&dedupFlags.test, "test", "t", false, "test")
	})
}
