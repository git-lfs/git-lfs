package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"path/filepath"
	"testing"
)

func TestFsck(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("fsck")
	cmd.Output = "Object a.dat (916f0027a575074ce72a331777c3478d6513f786a591bd892da1a577bf2335f9) is corrupt"

	testFileContent := "test data"
	h := sha256.New()
	io.WriteString(h, testFileContent)
	wantOid := hex.EncodeToString(h.Sum(nil))
	lfsObjectPath := filepath.Join(repo.Path, ".git", "lfs", "objects", wantOid[0:2], wantOid[2:4], wantOid)

	cmd.Before(func() {
		path := filepath.Join(repo.Path, ".git", "info", "attributes")
		repo.WriteFile(path, "*.dat filter=lfs -crlf\n")

		// Add a Git LFS object
		repo.WriteFile(filepath.Join(repo.Path, "a.dat"), testFileContent)
		repo.GitCmd("add", "a.dat")
		repo.GitCmd("commit", "-m", "a")
		repo.WriteFile(lfsObjectPath, testFileContent+"CORRUPTION")
	})
}
