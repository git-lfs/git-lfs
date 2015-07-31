package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	fsckDryRun bool

	fsckCmd = &cobra.Command{
		Use:   "fsck",
		Short: "Verifies validity of Git LFS files",
		Run:   fsckCommand,
	}
)

func doFsck() (bool, error) {
	ref, err := git.CurrentRef()
	if err != nil {
		return false, err
	}

	// The LFS scanner methods return unexported *lfs.wrappedPointer objects.
	// All we care about is the pointer OID and file name
	pointerIndex := make(map[string]string)

	pointers, err := lfs.ScanRefs(ref, "", nil)
	if err != nil {
		return false, err
	}

	for _, p := range pointers {
		pointerIndex[p.Oid] = p.Name
	}

	// TODO(zeroshirts): do we want to look for LFS stuff in past commits?
	p2, err := lfs.ScanIndex()
	if err != nil {
		return false, err
	}

	for _, p := range p2 {
		pointerIndex[p.Oid] = p.Name
	}

	ok := true

	for oid, name := range pointerIndex {
		path := filepath.Join(lfs.LocalMediaDir, oid[0:2], oid[2:4], oid)

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

			badDir := filepath.Join(lfs.LocalGitStorageDir, "lfs", "bad")
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
	fsckCmd.Flags().BoolVarP(&fsckDryRun, "dry-run", "d", false, "List corrupt objects without deleting them.")
	RootCmd.AddCommand(fsckCmd)
}
