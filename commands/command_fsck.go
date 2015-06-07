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

	pointers, err := lfs.ScanRefs(ref, "")
	if err != nil {
		return false, err
	}

	// TODO(zeroshirts): do we want to look for LFS stuff in past commits?
	p2, err := lfs.ScanIndex()
	if err != nil {
		return false, err
	}

	// zeroshirts: assuming no duplicates...
	pointers = append(pointers, p2...)

	ok := true

	for _, p := range pointers {
		path := filepath.Join(localGitDir, "lfs", "objects", p.Pointer.Oid[0:2], p.Pointer.Oid[2:4], p.Pointer.Oid)

		Debug("Examining %v (%v)", p.Name, path)

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
		if recalculatedOid != p.Pointer.Oid {
			ok = false
			Print("Object %s (%s) is corrupt", p.Name, p.Oid)
			os.RemoveAll(path)
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
	RootCmd.AddCommand(fsckCmd)
}
