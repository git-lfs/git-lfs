package githistory

import (
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git/odb"
	"github.com/stretchr/testify/assert"
)

func TestRewriterRewritesHistory(t *testing.T) {
	db := DatabaseFromFixture(t, "linear-history.git")
	r := NewRewriter(db)

	tip, err := r.Rewrite(&RewriteOptions{Include: []string{"refs/heads/master"},
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

func TestRewriterRewritesOctopusMerges(t *testing.T) {
	db := DatabaseFromFixture(t, "octopus-merge.git")
	r := NewRewriter(db)

	tip, err := r.Rewrite(&RewriteOptions{Include: []string{"refs/heads/master"},
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

func TestRewriterDoesntVisitUnchangedSubtrees(t *testing.T) {
	db := DatabaseFromFixture(t, "repeated-subtrees.git")
	r := NewRewriter(db)

	seen := make(map[string]int)

	_, err := r.Rewrite(&RewriteOptions{Include: []string{"refs/heads/master"},
		BlobFn: func(path string, b *odb.Blob) (*odb.Blob, error) {
			seen[path] = seen[path] + 1

			return b, nil
		},
	})

	assert.Nil(t, err)

	assert.Equal(t, 2, seen[root("a.txt")])
	assert.Equal(t, 1, seen[root(filepath.Join("subdir", "b.txt"))])
}

func TestRewriterVisitsUniqueEntriesWithIdenticalContents(t *testing.T) {
	db := DatabaseFromFixture(t, "identical-blobs.git")
	r := NewRewriter(db)

	tip, err := r.Rewrite(&RewriteOptions{Include: []string{"refs/heads/master"},
		BlobFn: func(path string, b *odb.Blob) (*odb.Blob, error) {
			if path == root("b.txt") {
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

func TestRewriterIgnoresPathsThatDontMatchFilter(t *testing.T) {
	include := []string{"*.txt"}
	exclude := []string{"subdir/**/*.txt"}

	filter := filepathfilter.New(include, exclude)

	db := DatabaseFromFixture(t, "non-repeated-subtrees.git")
	r := NewRewriter(db, WithFilter(filter))

	seen := make(map[string]int)

	_, err := r.Rewrite(&RewriteOptions{Include: []string{"refs/heads/master"},
		BlobFn: func(path string, b *odb.Blob) (*odb.Blob, error) {
			seen[path] = seen[path] + 1

			return b, nil
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, 1, seen[root("a.txt")])
	assert.Equal(t, 0, seen[root(filepath.Join("subdir", "b.txt"))])
}

func TestRewriterAllowsAdditionalTreeEntries(t *testing.T) {
	db := DatabaseFromFixture(t, "linear-history.git")
	r := NewRewriter(db)

	extra, err := db.WriteBlob(&odb.Blob{
		Contents: strings.NewReader("extra\n"),
		Size:     int64(len("extra\n")),
	})
	assert.Nil(t, err)

	tip, err := r.Rewrite(&RewriteOptions{Include: []string{"refs/heads/master"},
		BlobFn: func(path string, b *odb.Blob) (*odb.Blob, error) {
			return b, nil
		},

		TreeCallbackFn: func(path string, tr *odb.Tree) (*odb.Tree, error) {
			return &odb.Tree{
				Entries: append(tr.Entries, &odb.TreeEntry{
					Name:     "extra.txt",
					Filemode: 0100644,
					Oid:      extra,
				}),
			}, nil
		},
	})

	assert.Nil(t, err)

	tree1 := "40c2eb627a3b8e84b82a47a973d32960f3898b6a"
	tree2 := "d7a5bcb69f2cd2652a014663a948952ea603c2c0"
	tree3 := "45b752554d128f85bf23d7c3ddf48c47cbc345c8"

	// After rewriting, the HEAD state of the repository should contain a
	// tree identical to:
	//
	//   100644 blob e440e5c842586965a7fb77deda2eca68612b1f53    hello.txt
	//   100644 blob 0f2287157f7cb0dd40498c7a92f74b6975fa2d57    extra.txt

	AssertCommitTree(t, db, hex.EncodeToString(tip), tree1)

	AssertBlobContents(t, db, tree1, "hello.txt", "3")
	AssertBlobContents(t, db, tree1, "extra.txt", "extra\n")

	// After rewriting, the HEAD~1 state of the repository should contain a
	// tree identical to:
	//
	//   100644 blob d8263ee9860594d2806b0dfd1bfd17528b0ba2a4    hello.txt
	//   100644 blob 0f2287157f7cb0dd40498c7a92f74b6975fa2d57    extra.txt

	AssertCommitParent(t, db, hex.EncodeToString(tip), "45af5deb9a25bc4069b15c1f5bdccb0340978707")
	AssertCommitTree(t, db, "45af5deb9a25bc4069b15c1f5bdccb0340978707", tree2)

	AssertBlobContents(t, db, tree2, "hello.txt", "2")
	AssertBlobContents(t, db, tree2, "extra.txt", "extra\n")

	// After rewriting, the HEAD~2 state of the repository should contain a
	// tree identical to:
	//
	//   100644 blob 56a6051ca2b02b04ef92d5150c9ef600403cb1de    hello.txt
	//   100644 blob 0f2287157f7cb0dd40498c7a92f74b6975fa2d57    extra.txt

	AssertCommitParent(t, db, "45af5deb9a25bc4069b15c1f5bdccb0340978707", "99f6bd7cd69b45494afed95b026f3e450de8304f")
	AssertCommitTree(t, db, "99f6bd7cd69b45494afed95b026f3e450de8304f", tree3)

	AssertBlobContents(t, db, tree3, "hello.txt", "1")
	AssertBlobContents(t, db, tree3, "extra.txt", "extra\n")
}

func TestHistoryRewriterUseOriginalParentsForPartialMigration(t *testing.T) {
	db := DatabaseFromFixture(t, "linear-history-with-tags.git")
	r := NewRewriter(db)

	tip, err := r.Rewrite(&RewriteOptions{
		Include: []string{"refs/heads/master"},
		Exclude: []string{"refs/tags/middle"},

		BlobFn: func(path string, b *odb.Blob) (*odb.Blob, error) {
			return b, nil
		},
	})

	// After rewriting, the rewriter should have only modified the latest
	// commit (HEAD), and excluded the first two, both reachable by
	// refs/tags/middle.
	//
	// This should modify one commit, and appropriately link the parent as
	// follows:
	//
	//   tree 20ecedad3e74a113695fe5f00ab003694e2e1e9c
	//   parent 228afe30855933151f7a88e70d9d88314fd2f191
	//   author Taylor Blau <me@ttaylorr.com> 1496954214 -0600
	//   committer Taylor Blau <me@ttaylorr.com> 1496954214 -0600
	//
	//   some.txt: c

	expectedParent := "228afe30855933151f7a88e70d9d88314fd2f191"

	assert.NoError(t, err)
	AssertCommitParent(t, db, hex.EncodeToString(tip), expectedParent)
}

func root(path string) string {
	if !strings.HasPrefix(path, string(os.PathSeparator)) {
		path = string(os.PathSeparator) + path
	}
	return path
}
