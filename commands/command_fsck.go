package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
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
	lfs.InstallHooks(false)
	requireInRepo()

	ref, err := git.CurrentRef()
	if err != nil {
		ExitWithError(err)
	}

	var corruptOids []string
	gitscanner := lfs.NewGitScanner(func(p *lfs.WrappedPointer, err error) {
		if err == nil {
			var pointerOk bool
			pointerOk, err = fsckPointer(p.Name, p.Oid)
			if !pointerOk {
				corruptOids = append(corruptOids, p.Oid)
			}
		}

		if err != nil {
			Panic(err, "Error checking Git LFS files")
		}
	})

	if err := gitscanner.ScanRef(ref.Sha, nil); err != nil {
		ExitWithError(err)
	}

	if err := gitscanner.ScanIndex("HEAD", nil); err != nil {
		ExitWithError(err)
	}

	gitscanner.Close()

	if len(corruptOids) == 0 {
		Print("Git LFS fsck OK")
		return
	}

	if fsckDryRun {
		return
	}

	storageConfig := config.Config.StorageConfig()
	badDir := filepath.Join(storageConfig.LfsStorageDir, "bad")
	Print("Moving corrupt objects to %s", badDir)

	if err := os.MkdirAll(badDir, 0755); err != nil {
		ExitWithError(err)
	}

	for _, oid := range corruptOids {
		badFile := filepath.Join(badDir, oid)
		if err := os.Rename(lfs.LocalMediaPathReadOnly(oid), badFile); err != nil {
			ExitWithError(err)
		}
	}
}

func fsckPointer(name, oid string) (bool, error) {
	path := lfs.LocalMediaPathReadOnly(oid)

	Debug("Examining %v (%v)", name, path)

	f, err := os.Open(path)
	if pErr, pOk := err.(*os.PathError); pOk {
		Print("Object %s (%s) could not be checked: %s", name, oid, pErr.Err)
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

	Print("Object %s (%s) is corrupt", name, oid)
	return false, nil
}

func init() {
	RegisterCommand("fsck", fsckCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&fsckDryRun, "dry-run", "d", false, "List corrupt objects without deleting them.")
	})
}
