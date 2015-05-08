package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
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
	oid1 := hex.EncodeToString(h.Sum(nil))
	lfsObjectPath := filepath.Join(repo.Path, ".git", "lfs", "objects", oid1[0:2], oid1[2:4], oid1)

	testFile2Content := "test data 2"
	h.Reset()
	io.WriteString(h, testFile2Content)
	oid2 := hex.EncodeToString(h.Sum(nil))
	lfsObject2Path := filepath.Join(repo.Path, ".git", "lfs", "objects", oid2[0:2], oid2[2:4], oid2)

	cmd.Before(func() {
		path := filepath.Join(repo.Path, ".git", "info", "attributes")
		repo.WriteFile(path, "*.dat filter=lfs -crlf\n")

		// Add a Git LFS object
		repo.WriteFile(filepath.Join(repo.Path, "a.dat"), testFileContent)
		repo.WriteFile(filepath.Join(repo.Path, "b.dat"), testFile2Content)
		repo.GitCmd("add", "*.dat")
		repo.GitCmd("commit", "-m", "a")
		repo.WriteFile(lfsObjectPath, testFileContent+"CORRUPTION")
	})

	cmd.After(func() {
		by, err := ioutil.ReadFile(lfsObject2Path)
		if err != nil {
			t.Fatal(err)
		}

		h.Reset()
		h.Write(by)
		oid := hex.EncodeToString(h.Sum(nil))
		if oid != oid2 {
			t.Errorf("oid for b.dat does not match")
		}

		_, err = os.Stat(lfsObjectPath)
		if err == nil {
			t.Errorf("Expected a.dat to be cleared for being corrupt", lfsObjectPath)
		}
	})
}
