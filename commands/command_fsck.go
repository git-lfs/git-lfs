package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
)

var (
	fsckDryRun bool

	fsckCmd = &cobra.Command{
		Use:   "fsck",
		Short: "Verifies validity of Git LFS files",
		Run:   fsckCommand,
	}
)

func doFsck(localGitDir string) (bool, error) {
	ref, err := git.CurrentRef()
	if err != nil {
		return false, err
	}

	// The LFS scanner methods return unexported *lfs.wrappedPointer objects.
	// All we care about is the pointer OID and file name
	pointerIndex := make(map[string]string)

	pointers, err := lfs.ScanRefs(ref, "")
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
		path := filepath.Join(localGitDir, "lfs", "objects", oid[0:2], oid[2:4], oid)

		Debug("Examining %v (%v)", name, path)

		f, err := os.Open(path)
		if err != nil {
			return false, err
		}
		defer f.Close()

		oidHash := sha256.New()
		_, err = io.Copy(oidHash, f)
		if err != nil {
			return false, err
		}

		recalculatedOid := hex.EncodeToString(oidHash.Sum(nil))
		if recalculatedOid != oid {
			ok = false
			Print("Object %s (%s) is corrupt", name, oid)
			if !fsckDryRun {
				os.RemoveAll(path)
			}
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

	ok, err := doFsck(lfs.LocalGitDir)
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
