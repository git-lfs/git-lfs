package githistory

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
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

	AssertCommitParent(t, db, hex.EncodeToString(tip), "911994ab82ce256433c1fa739dbbbc7142156289")
	AssertCommitTree(t, db, "911994ab82ce256433c1fa739dbbbc7142156289", tree2)

	AssertBlobContents(t, db, tree2, "hello.txt", "3")

	// After rewriting, the HEAD~2 state of the repository should contain a
	// tree identical to:
	//
	//   100644 blob d8263ee9860594d2806b0dfd1bfd17528b0ba2a4   hello.txt

	AssertCommitParent(t, db, "911994ab82ce256433c1fa739dbbbc7142156289", "38679ebeba3403103196eb6272b326f96c928ace")
	AssertCommitTree(t, db, "38679ebeba3403103196eb6272b326f96c928ace", tree3)

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

	AssertCommitParent(t, db, hex.EncodeToString(tip), "89ab88fb7e11a439299aa2aa77a5d98f6629b750")
	AssertCommitParent(t, db, hex.EncodeToString(tip), "adf1e9085f9dd263c1bec399b995ccfa5d994721")

	// And each of those parents should contain the root commit as their own
	// parent:

	AssertCommitParent(t, db, "89ab88fb7e11a439299aa2aa77a5d98f6629b750", "52daca68bcf750bb86289fd95f92f5b3bd202328")
	AssertCommitParent(t, db, "adf1e9085f9dd263c1bec399b995ccfa5d994721", "52daca68bcf750bb86289fd95f92f5b3bd202328")
}

func TestRewriterVisitsPackedObjects(t *testing.T) {
	db := DatabaseFromFixture(t, "packed-objects.git")
	r := NewRewriter(db)

	var contents []byte

	_, err := r.Rewrite(&RewriteOptions{Include: []string{"refs/heads/master"},
		BlobFn: func(path string, b *odb.Blob) (*odb.Blob, error) {
			var err error

			contents, err = ioutil.ReadAll(b.Contents)
			if err != nil {
				return nil, err
			}

			return &odb.Blob{
				Contents: bytes.NewReader(contents),
				Size:     int64(len(contents)),
			}, nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, string(contents), "Hello, world!\n")
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

	assert.Equal(t, 2, seen["a.txt"])
	assert.Equal(t, 1, seen["subdir/b.txt"])
}

func TestRewriterVisitsUniqueEntriesWithIdenticalContents(t *testing.T) {
	db := DatabaseFromFixture(t, "identical-blobs.git")
	r := NewRewriter(db)

	tip, err := r.Rewrite(&RewriteOptions{Include: []string{"refs/heads/master"},
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

	fmt.Println(hex.EncodeToString(tip))
	root, _ := db.Root()
	fmt.Println(root)

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
	exclude := []string{"subdir/*.txt"}

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
	assert.Equal(t, 1, seen["a.txt"])
	assert.Equal(t, 0, seen["subdir/b.txt"])
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

	AssertCommitParent(t, db, hex.EncodeToString(tip), "54ca0fdd5ee455d872ce4b4e379abe1c4cdc39b3")
	AssertCommitTree(t, db, "54ca0fdd5ee455d872ce4b4e379abe1c4cdc39b3", tree2)

	AssertBlobContents(t, db, tree2, "hello.txt", "2")
	AssertBlobContents(t, db, tree2, "extra.txt", "extra\n")

	// After rewriting, the HEAD~2 state of the repository should contain a
	// tree identical to:
	//
	//   100644 blob 56a6051ca2b02b04ef92d5150c9ef600403cb1de    hello.txt
	//   100644 blob 0f2287157f7cb0dd40498c7a92f74b6975fa2d57    extra.txt

	AssertCommitParent(t, db, "54ca0fdd5ee455d872ce4b4e379abe1c4cdc39b3", "4c52196256c611d18ad718b9b68b3d54d0a6686d")
	AssertCommitTree(t, db, "4c52196256c611d18ad718b9b68b3d54d0a6686d", tree3)

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

func TestHistoryRewriterUpdatesRefs(t *testing.T) {
	db := DatabaseFromFixture(t, "linear-history.git")
	r := NewRewriter(db)

	AssertRef(t, db,
		"refs/heads/master", HexDecode(t, "e669b63f829bfb0b91fc52a5bcea53dd7977a0ee"))

	tip, err := r.Rewrite(&RewriteOptions{
		Include: []string{"refs/heads/master"},

		UpdateRefs: true,

		BlobFn: func(path string, b *odb.Blob) (*odb.Blob, error) {
			suffix := strings.NewReader("_suffix")

			return &odb.Blob{
				Contents: io.MultiReader(b.Contents, suffix),
				Size:     b.Size + int64(suffix.Len()),
			}, nil
		},
	})

	assert.Nil(t, err)

	c1 := hex.EncodeToString(tip)
	c2 := "86f7ba8f02edaca4f980cdd584ea8899e18b840c"
	c3 := "d73b8c1a294e2371b287d9b75dbed82328ad446e"

	AssertRef(t, db, "refs/heads/master", tip)

	AssertCommitParent(t, db, c1, c2)
	AssertCommitParent(t, db, c2, c3)
}

func TestHistoryRewriterReturnsFilter(t *testing.T) {
	f := filepathfilter.New([]string{"a"}, []string{"b"})
	r := NewRewriter(nil, WithFilter(f))

	expected := reflect.ValueOf(f).Elem().Addr().Pointer()
	got := reflect.ValueOf(r.Filter()).Elem().Addr().Pointer()

	assert.Equal(t, expected, got,
		"git/githistory: expected Rewriter.Filter() to return same *filepathfilter.Filter instance")
}
