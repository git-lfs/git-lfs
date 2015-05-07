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

type fsckError struct {
	name, oid string
}

func (e *fsckError) Error() string {
	return "Object " + e.name + " (" + e.oid + ") is corrupt"
}

var (
	fsckCmd = &cobra.Command{
		Use:   "fsck",
		Short: "Verifies validity of Git LFS files",
		Run:   fsckCommand,
	}
)

func doFsck(localGitDir string) error {
	ref, err := git.CurrentRef()
	if err != nil {
		return err
	}

	pointers, err := lfs.ScanRefs(ref, "")
	if err != nil {
		return err
	}

	// TODO(zeroshirts): do we want to look for LFS stuff in past commits?
	p2, err := lfs.ScanIndex()
	if err != nil {
		return err
	}
	// zeroshirts: assuming no duplicates...
	pointers = append(pointers, p2...)

	for _, p := range pointers {
		path := filepath.Join(localGitDir, "lfs", "objects", p.Pointer.Oid[0:2], p.Pointer.Oid[2:4], p.Pointer.Oid)

		Debug("Examining %v (%v)", p.Name, path)

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		oidHash := sha256.New()
		_, err = io.Copy(oidHash, f)
		if err != nil {
			return err
		}
		f.Close()

		recalculatedOid := hex.EncodeToString(oidHash.Sum(nil))
		if recalculatedOid != p.Pointer.Oid {
			return &fsckError{p.Name, p.Pointer.Oid}
		}
		Debug("%v (%v) intact", p.Name, path)
	}
	return nil
}

// TODO(zeroshirts): 'git fsck' reports status (percentage, current#/total) as
// it checks... we should do the same, as we are rehashing potentially gigs and
// gigs of content.
//
// NOTE(zeroshirts): Ideally git would have hooks for fsck such that we could
// chain a lfs-fsck, but I don't think it does.
func fsckCommand(cmd *cobra.Command, args []string) {
	lfs.InstallHooks(false)

	err := doFsck(lfs.LocalGitDir)
	if err != nil {
		Panic(err, "Git LFS fsck failed")
	}
	Print("Git LFS fsck OK")
}

func init() {
	RootCmd.AddCommand(fsckCmd)
}
