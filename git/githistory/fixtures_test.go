package githistory

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/git-lfs/git-lfs/git/odb"
	"github.com/stretchr/testify/assert"
)

// DatabaseFromFixture returns a *git/odb.ObjectDatabase instance that is safely
// mutable and created from a template equivelant to the fixture that you
// provided it.
//
// If any error was encountered, it will call t.Fatalf() immediately.
func DatabaseFromFixture(t *testing.T, name string) *odb.ObjectDatabase {
	path, err := copyToTmp(filepath.Join("fixtures", name))
	if err != nil {
		t.Fatalf("git/odb: could not copy fixture %s: %v", name, err)
	}

	db, err := odb.FromFilesystem(filepath.Join(path, "objects"))
	if err != nil {
		t.Fatalf("git/odb: could not create object database: %v", err)
	}
	return db
}

// AssertBlobContents asserts that the blob contents given by loading the path
// starting from the root tree "tree" has the given "contents".
func AssertBlobContents(t *testing.T, db *odb.ObjectDatabase, tree, path, contents string) {
	// First, load the root tree.
	root, err := db.Tree(HexDecode(t, tree))
	if err != nil {
		t.Fatalf("git/odb: cannot load tree: %s: %s", tree, err)
	}

	// Then, iterating through each part of the filepath (i.e., a/b/c.txt ->
	// []string{"a", "b", "c.txt"}).
	parts := strings.Split(path, string(os.PathSeparator))
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]

		// Load the subtree given by that name.
		var subtree *odb.Tree
		for _, entry := range root.Entries {
			if entry.Name != part {
				continue
			}

			subtree, err = db.Tree(entry.Oid)
			if err != nil {
				t.Fatalf("git/odb: cannot load subtree %s: %s", filepath.Join(parts[:i]...), err)
			}
			break
		}

		if subtree == nil {
			t.Fatalf("git/odb: subtree %s does not exist", path)
		}

		// And re-assign it to root, creating a sort of pseudo-recursion.
		root = subtree
	}

	filename := parts[len(parts)-1]

	// Find the blob given by the last entry in parts (the filename).
	var blob *odb.Blob
	for _, entry := range root.Entries {
		if entry.Name == filename {
			blob, err = db.Blob(entry.Oid)
			if err != nil {
				t.Fatalf("git/odb: cannot load blob %x: %s", entry.Oid, err)
			}
		}
	}

	// If we couldn't find the blob, fail immediately.
	if blob == nil {
		t.Fatalf("git/odb: blob at %s in %s does not exist", path, tree)
	}

	// Perform an assertion on the blob's contents.
	got, err := ioutil.ReadAll(blob.Contents)
	if err != nil {
		t.Fatalf("git/odb: cannot read contents from blob %s: %s", path, err)
	}

	assert.Equal(t, contents, string(got))
}

// AssertCommitParent asserts that the given commit has a parent equivalent to
// the one provided.
func AssertCommitParent(t *testing.T, db *odb.ObjectDatabase, sha, parent string) {
	commit, err := db.Commit(HexDecode(t, sha))
	if err != nil {
		t.Fatalf("git/odb: expected to read commit: %s, couldn't: %v", sha, err)
	}

	decoded, err := hex.DecodeString(parent)
	if err != nil {
		t.Fatalf("git/odb: expected to decode parent SHA: %s, couldn't: %v", parent, err)
	}

	assert.Contains(t, commit.ParentIDs, decoded,
		"git/odb: expected parents of commit: %s to contain: %s", sha, parent)
}

// AssertCommitTree asserts that the given commit has a tree equivelant to the
// one provided.
func AssertCommitTree(t *testing.T, db *odb.ObjectDatabase, sha, tree string) {
	commit, err := db.Commit(HexDecode(t, sha))
	if err != nil {
		t.Fatalf("git/odb: expected to read commit: %s, couldn't: %v", sha, err)
	}

	decoded, err := hex.DecodeString(tree)
	if err != nil {
		t.Fatalf("git/odb: expected to decode tree SHA: %s, couldn't: %v", tree, err)
	}

	assert.Equal(t, decoded, commit.TreeID, "git/odb: expected tree ID: %s (got: %x)", tree, commit.TreeID)
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
