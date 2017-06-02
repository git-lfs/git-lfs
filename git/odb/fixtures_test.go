package odb

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// DatabaseFromFixture returns a *git/odb.ObjectDatabase instance that is safely
// mutable and created from a template equivelant to the fixture that you
// provided it.
//
// If any error was encountered, it will call t.Fatalf() immediately.
func DatabaseFromFixture(t *testing.T, name string) *ObjectDatabase {
	path, err := copyToTmp(filepath.Join("fixtures", name))
	if err != nil {
		t.Fatalf("git/odb: could not copy fixture %s: %v", name, err)
	}

	db, err := FromFilesystem(filepath.Join(path, "objects"))
	if err != nil {
		t.Fatalf("git/odb: could not create object database: %v", err)
	}
	return db
}

// HexDecode decodes the given ASCII hex-encoded string into []byte's, or fails
// the test immediately if the given "sha" wasn't a valid hex-encoded sequence.
func HexDecode(t *testing.T, sha string) []byte {
	b, err := hex.DecodeString(sha)
	if err != nil {
		t.Fatalf("git/odb: could not decode string: %q, %v", sha, err)
	}

	return b
}

// copyToTmp copies the given fixutre to a folder in /tmp.
func copyToTmp(fixture string) (string, error) {
	p, err := ioutil.TempDir("", fmt.Sprintf("git-lfs-fixture-%s", filepath.Dir(fixture)))
	if err != nil {
		return "", err
	}

	if err = copyDir(fixture, p); err != nil {
		return "", err
	}
	return p, nil
}

// copyDir copies a directory (and recursively all files and subdirectories)
// from "from" to "to" preserving permissions and ownership.
func copyDir(from, to string) error {
	stat, err := os.Stat(from)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(to, stat.Mode()); err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(from)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sp := filepath.Join(from, entry.Name())
		dp := filepath.Join(to, entry.Name())

		if entry.IsDir() {
			err = copyDir(sp, dp)
		} else {
			err = copyFile(sp, dp)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

// copyFile copies a file from "from" to "to" preserving permissions and
// ownership.
func copyFile(from, to string) error {
	src, err := os.Open(from)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(to)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	stat, err := os.Stat(from)
	if err != nil {
		return err
	}

	return os.Chmod(to, stat.Mode())
}
