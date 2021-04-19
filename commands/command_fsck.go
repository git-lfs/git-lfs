package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/spf13/cobra"
)

var (
	fsckDryRun bool
)

// TODO(zeroshirts): 'git fsck' reports status (percentage, current#/total) as
// it checks... we should do the same, as we are rehashing potentially gigs and
// gigs of content.
//
// NOTE(zeroshirts): Ideally git would have hooks for fsck such that we could
// chain a lfs-fsck, but I don't think it does.
func fsckCommand(cmd *cobra.Command, args []string) {
	installHooks(false)
	setupRepository()

	ref, err := git.CurrentRef()
	if err != nil {
		ExitWithError(err)
	}

	ok := true
	var corruptOids []string
	corruptOids = doFsckObjects(ref)
	ok = ok && corruptOids == nil

	if ok {
		Print("Git LFS fsck OK")
		return
	}

	if fsckDryRun {
		return
	}

	if len(corruptOids) != 0 {
		badDir := filepath.Join(cfg.LFSStorageDir(), "bad")
		Print("objects: repair: moving corrupt objects to %s", badDir)

		if err := tools.MkdirAll(badDir, cfg); err != nil {
			ExitWithError(err)
		}

		for _, oid := range corruptOids {
			badFile := filepath.Join(badDir, oid)
			if err := os.Rename(cfg.Filesystem().ObjectPathname(oid), badFile); err != nil {
				ExitWithError(err)
			}
		}
	}
}

// doFsckObjects checks that the objects in the given ref are correct and exist.
func doFsckObjects(ref *git.Ref) []string {
	var corruptOids []string
	gitscanner := lfs.NewGitScanner(cfg, func(p *lfs.WrappedPointer, err error) {
		if err == nil {
			var pointerOk bool
			pointerOk, err = fsckPointer(p.Name, p.Oid, p.Size)
			if !pointerOk {
				corruptOids = append(corruptOids, p.Oid)
			}
		}

		if err != nil {
			Panic(err, "Error checking Git LFS files")
		}
	})

	// If 'lfs.fetchexclude' is set and 'git lfs fsck' is run after the
	// initial fetch (i.e., has elected to fetch a subset of Git LFS
	// objects), the "missing" ones will fail the fsck.
	//
	// Attach a filepathfilter to avoid _only_ the excluded paths.
	gitscanner.Filter = filepathfilter.New(nil, cfg.FetchExcludePaths())

	if err := gitscanner.ScanRef(ref.Sha, nil); err != nil {
		ExitWithError(err)
	}

	if err := gitscanner.ScanIndex("HEAD", nil); err != nil {
		ExitWithError(err)
	}

	gitscanner.Close()
	return corruptOids
}

func fsckPointer(name, oid string, size int64) (bool, error) {
	path := cfg.Filesystem().ObjectPathname(oid)

	Debug("Examining %v (%v)", name, path)

	f, err := os.Open(path)
	if pErr, pOk := err.(*os.PathError); pOk {
		// This is an empty file.  No problem here.
		if size == 0 {
			return true, nil
		}
		Print("objects: openError: %s (%s) could not be checked: %s", name, oid, pErr.Err)
		return false, nil
	}

	if err != nil {
		return false, err
	}

	oidHash := sha256.New()
	_, err = io.Copy(oidHash, f)
	f.Close()
	if err != nil {
		return false, err
	}

	recalculatedOid := hex.EncodeToString(oidHash.Sum(nil))
	if recalculatedOid == oid {
		return true, nil
	}

	Print("objects: corruptObject: %s (%s) is corrupt", name, oid)
	return false, nil
}

func init() {
	RegisterCommand("fsck", fsckCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&fsckDryRun, "dry-run", "d", false, "List corrupt objects without deleting them.")
	})
}
