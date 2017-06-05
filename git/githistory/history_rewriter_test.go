package githistory

import (
	"encoding/hex"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/git-lfs/git-lfs/git/odb"
	"github.com/stretchr/testify/assert"
)

func TestHistoryRewriterRewritesHistory(t *testing.T) {
	db := DatabaseFromFixture(t, "linear-history.git")
	r := NewHistoryRewriter(db)

	tip, err := r.Rewrite(&RewriteOptions{Left: "master",
		BlobFn: func(path string, b *odb.Blob) (*odb.Blob, error) {
			contents, err := ioutil.ReadAll(b.Contents)
			if err != nil {
				return nil, err
			}

			n, err := strconv.Atoi(string(contents))
			if err != nil {
				return nil, err
			}

			rewritten := strconv.Itoa(n + 1)

			return &odb.Blob{
				Contents: strings.NewReader(rewritten),
				Size:     int64(len(rewritten)),
			}, nil
		},
	})

	assert.Nil(t, err)

	tree1 := "ad0aebd16e34cf047820994ea7538a6d4a111082"
	tree2 := "6e07bd31cb70c4add2c973481ad4fa38b235ca69"
	tree3 := "c5decfe1fcf39b8c489f4a0bf3b3823676339f80"

	// After rewriting, the HEAD state of the repository should contain a
	// tree identical to:
	//
	//   100644 blob bf0d87ab1b2b0ec1a11a3973d2845b42413d9767   hello.txt

	AssertCommitTree(t, db, hex.EncodeToString(tip), tree1)

	AssertBlobContents(t, db, tree1, "hello.txt", "4")

	// After rewriting, the HEAD~1 state of the repository should contain a
	// tree identical to:
	//
	//   100644 blob e440e5c842586965a7fb77deda2eca68612b1f53   hello.txt

	AssertCommitParent(t, db, hex.EncodeToString(tip), "4aaa3f49ffeabbb874250fe13ffeb8c683aba650")
	AssertCommitTree(t, db, "4aaa3f49ffeabbb874250fe13ffeb8c683aba650", tree2)

	AssertBlobContents(t, db, tree2, "hello.txt", "3")

	// After rewriting, the HEAD~2 state of the repository should contain a
	// tree identical to:
	//
	//   100644 blob d8263ee9860594d2806b0dfd1bfd17528b0ba2a4   hello.txt

	AssertCommitParent(t, db, "4aaa3f49ffeabbb874250fe13ffeb8c683aba650", "24a341e1ff75addc22e336a8d87f82ba56b86fcf")
	AssertCommitTree(t, db, "24a341e1ff75addc22e336a8d87f82ba56b86fcf", tree3)

	AssertBlobContents(t, db, tree3, "hello.txt", "2")
}

func TestHistoryRewriterRewritesOctopusMerges(t *testing.T) {
	db := DatabaseFromFixture(t, "octopus-merge.git")
	r := NewHistoryRewriter(db)

	tip, err := r.Rewrite(&RewriteOptions{Left: "master",
		BlobFn: func(path string, b *odb.Blob) (*odb.Blob, error) {
			return &odb.Blob{
				Contents: io.MultiReader(b.Contents, strings.NewReader("_new")),
				Size:     b.Size + int64(len("_new")),
			}, nil
		},
	})

	assert.Nil(t, err)

	tree := "8a56716daa78325c3d0433cc163890969810b0da"

	// After rewriting, the HEAD state of the repository should contain a
	// tree identical to:
	//
	//   100644 blob 309f7fc2bfd9ae77b4131cf9cbcc3b548c42ca57    a.txt
	//   100644 blob 70470dc26cb3eef54fe3dcba53066f7ca7c495c0    b.txt
	//   100644 blob f2557f74fd5b60f959baf77091782089761e2dc3    hello.txt

	AssertCommitTree(t, db, hex.EncodeToString(tip), tree)

	AssertBlobContents(t, db, tree, "a.txt", "a_new")
	AssertBlobContents(t, db, tree, "b.txt", "b_new")
	AssertBlobContents(t, db, tree, "hello.txt", "hello_new")

	// And should contain the following parents:
	//
	//   parent 1fe2b9577d5610e8d8fb2c3030534036fb648393
	//   parent ca447959bdcd20253d69b227bcc7c2e1d3126d5c

	AssertCommitParent(t, db, hex.EncodeToString(tip), "1fe2b9577d5610e8d8fb2c3030534036fb648393")
	AssertCommitParent(t, db, hex.EncodeToString(tip), "ca447959bdcd20253d69b227bcc7c2e1d3126d5c")

	// And each of those parents should contain the root commit as their own
	// parent:

	AssertCommitParent(t, db, "1fe2b9577d5610e8d8fb2c3030534036fb648393", "9237567f379b3c83ddf53ad9a2ae3755afb62a09")
	AssertCommitParent(t, db, "ca447959bdcd20253d69b227bcc7c2e1d3126d5c", "9237567f379b3c83ddf53ad9a2ae3755afb62a09")
}

func TestHistoryRewriterDoesntVisitUnchangedSubtrees(t *testing.T) {
	db := DatabaseFromFixture(t, "repeated-subtrees.git")
	r := NewHistoryRewriter(db)

	seen := make(map[string]int)

	_, err := r.Rewrite(&RewriteOptions{Left: "master",
		BlobFn: func(path string, b *odb.Blob) (*odb.Blob, error) {
			seen[path] = seen[path] + 1

			return b, nil
		},
	})

	assert.Nil(t, err)

	assert.Equal(t, 2, seen["a.txt"])
	assert.Equal(t, 1, seen[filepath.Join("subdir", "b.txt")])
}

func TestHistoryRewriterVisitsUniqueEntriesWithIdenticalContents(t *testing.T) {
	db := DatabaseFromFixture(t, "identical-blobs.git")
	r := NewHistoryRewriter(db)

	tip, err := r.Rewrite(&RewriteOptions{Left: "master",
		BlobFn: func(path string, b *odb.Blob) (*odb.Blob, error) {
			if path == "b.txt" {
				return b, nil
			}

			return &odb.Blob{
				Contents: strings.NewReader("changed"),
				Size:     int64(len("changed")),
			}, nil
		},
	})

	assert.Nil(t, err)

	tree := "bbbe0a7676523ae02234bfe874784ca2380c2d4b"

	AssertCommitTree(t, db, hex.EncodeToString(tip), tree)

	// After rewriting, the HEAD state of the repository should contain a
	// tree identical to:
	//
	//   100644 blob 21fb1eca31e64cd3914025058b21992ab76edcf9    a.txt
	//   100644 blob 94f3610c08588440112ed977376f26a8fba169b0    b.txt

	AssertBlobContents(t, db, tree, "a.txt", "changed")
	AssertBlobContents(t, db, tree, "b.txt", "original")
}
