package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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

func doFsck() (bool, error) {
	requireInRepo()

	ref, err := git.CurrentRef()
	if err != nil {
		return false, err
	}

	// The LFS scanner methods return unexported *lfs.wrappedPointer objects.
	// All we care about is the pointer OID and file name
	pointerIndex, err := getPointersFromRef(ref.Sha)
	if err != nil {
		return false, err
	}

	ok := true

	for oid, name := range pointerIndex {
		path := lfs.LocalMediaPathReadOnly(oid)

		Debug("Examining %v (%v)", name, path)

		f, err := os.Open(path)
		if pErr, pOk := err.(*os.PathError); pOk {
			Print("Object %s (%s) could not be checked: %s", name, oid, pErr.Err)
			ok = false
			continue
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
		if recalculatedOid != oid {
			ok = false
			Print("Object %s (%s) is corrupt", name, oid)
			if fsckDryRun {
				continue
			}

			badDir := filepath.Join(config.LocalGitStorageDir, "lfs", "bad")
			if err := os.MkdirAll(badDir, 0755); err != nil {
				return false, err
			}

			badFile := filepath.Join(badDir, oid)
			if err := os.Rename(path, badFile); err != nil {
				return false, err
			}
			Print("  moved to %s", badFile)
		}
	}
	return ok, nil
}

func getPointersFromRef(ref string) (map[string]string, error) {
	// The LFS scanner methods return unexported *lfs.wrappedPointer objects.
	// All we care about is the pointer OID and file name
	pointerIndex := make(map[string]string)

	var multiErr error
	gitscanner := lfs.NewGitScanner(func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			if multiErr != nil {
				multiErr = fmt.Errorf("%v\n%v", multiErr, err)
			} else {
				multiErr = err
			}
			return
		}
		pointerIndex[p.Oid] = p.Name
	})

	defer gitscanner.Close()
	pointerCh, err := gitscanner.ScanRefWithDeleted(ref)
	if err != nil {
		return pointerIndex, err
	}

	for p := range pointerCh.Results {
		pointerIndex[p.Oid] = p.Name
	}

	if err := pointerCh.Wait(); err != nil {
		return pointerIndex, err
	}

	if err := gitscanner.ScanIndex("HEAD"); err != nil {
		return pointerIndex, err
	}

	return pointerIndex, multiErr
}

// TODO(zeroshirts): 'git fsck' reports status (percentage, current#/total) as
// it checks... we should do the same, as we are rehashing potentially gigs and
// gigs of content.
//
// NOTE(zeroshirts): Ideally git would have hooks for fsck such that we could
// chain a lfs-fsck, but I don't think it does.
func fsckCommand(cmd *cobra.Command, args []string) {
	lfs.InstallHooks(false)

	ok, err := doFsck()
	if err != nil {
		Panic(err, "Error checking Git LFS files")
	}

	if ok {
		Print("Git LFS fsck OK")
	}
}

func init() {
	RegisterCommand("fsck", fsckCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&fsckDryRun, "dry-run", "d", false, "List corrupt objects without deleting them.")
	})
}
