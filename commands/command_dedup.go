package commands

import (
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tr"
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
			err = errors.New(tr.Tr.Get("Unknown reason"))
		}
		Exit(tr.Tr.Get("This system does not support de-duplication: %s", err))
	}

	if len(cfg.Extensions()) > 0 {
		Exit(tr.Tr.Get("This platform supports file de-duplication, however, Git LFS extensions are configured and therefore de-duplication can not be used."))
	}

	Print(tr.Tr.Get("OK: This platform and repository support file de-duplication."))
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
		Exit(tr.Tr.Get("This system does not support de-duplication."))
	}

	if len(cfg.Extensions()) > 0 {
		Exit(tr.Tr.Get("This platform supports file de-duplication, however, Git LFS extensions are configured and therefore de-duplication can not be used."))
	}

	if dirty, err := git.IsWorkingCopyDirty(); err != nil {
		ExitWithError(err)
	} else if dirty {
		Exit(tr.Tr.Get("Working tree is dirty. Please commit or reset your change."))
	}

	// We assume working tree is clean.
	gitScanner := lfs.NewGitScanner(config.New(), func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			Exit(tr.Tr.Get("Could not scan for Git LFS tree: %s", err))
			return
		}

		if success, err := dedup(p); err != nil {
			// TRANSLATORS: Leading spaces should be included on
			// the second line so the format specifier aligns with
			// with the first format specifier on the first line.
			Error(tr.Tr.Get("Skipped: %s (Size: %d)\n          %s", p.Name, p.Size, err))
		} else if !success {
			Error(tr.Tr.Get("Skipped: %s (Size: %d)", p.Name, p.Size))
		} else if success {
			Print(tr.Tr.Get("Success: %s (Size: %d)", p.Name, p.Size))

			atomic.AddInt64(&dedupStats.totalProcessedCount, 1)
			atomic.AddInt64(&dedupStats.totalProcessedSize, p.Size)
		}
	})

	if err := gitScanner.ScanTree("HEAD", nil); err != nil {
		ExitWithError(err)
	}

	// TRANSLATORS: The second and third strings should have the colons
	// aligned in a column.
	Print("\n\n%s\n  %s\n  %s", tr.Tr.Get("Finished successfully."),
		tr.Tr.GetN(
			"De-duplicated  size: %d byte",
			"De-duplicated  size: %d bytes",
			int(dedupStats.totalProcessedSize),
			dedupStats.totalProcessedSize),
		tr.Tr.Get("              count: %d", dedupStats.totalProcessedCount))
}

// dedup executes
// Precondition: working tree MUST clean. We can replace working tree files from mediafile safely.
func dedup(p *lfs.WrappedPointer) (success bool, err error) {
	// PRECONDITION, check ofs object exists or skip this file.
	if !cfg.LFSObjectExists(p.Oid, p.Size) { // Not exists,
		// Basically, this is not happens because executing 'git status' in `git.IsWorkingCopyDirty()` recover it.
		return false, errors.New(tr.Tr.Get("Git LFS object file does not exist"))
	}

	// DO de-dup
	// Gather original state
	originalStat, err := os.Stat(p.Name)
	if err != nil {
		return false, err
	}

	// Do clone
	srcFile := cfg.Filesystem().ObjectPathname(p.Oid)
	if srcFile == os.DevNull {
		return true, nil
	}

	dstFile := filepath.Join(cfg.LocalWorkingDir(), p.Name)

	// Clone the file. This overwrites the destination if it exists.
	if ok, err := tools.CloneFileByPath(dstFile, srcFile); err != nil {
		return false, err
	} else if !ok {
		return false, errors.Errorf(tr.Tr.Get("unknown clone file error"))
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
