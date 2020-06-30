package gitattr

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/git-lfs/gitobj/v2"
	"github.com/git-lfs/wildmatch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	dat = wildmatch.NewWildmatch("*.dat",
		wildmatch.Basename,
		wildmatch.SystemCase)

	example = &Tree{
		Lines: []*Line{{
			Pattern: dat,
			Attrs: []*Attr{
				{
					K: "filter", V: "lfs",
				},
				{
					K: "diff", V: "lfs",
				},
				{
					K: "merge", V: "lfs",
				},
				{
					K: "text", V: "false",
				},
			},
		}},
		Children: map[string]*Tree{
			"subdir": &Tree{
				Lines: []*Line{{
					Pattern: dat,
					Attrs: []*Attr{
						{
							K: "subdir", V: "yes",
						},
					},
				}},
			},
		},
	}
)

func TestTreeAppliedInRoot(t *testing.T) {
	attrs := example.Applied("a.dat")

	assert.Len(t, attrs, 4)
	assert.Equal(t, attrs[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, attrs[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, attrs[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, attrs[3], &Attr{K: "text", V: "false"})
}

func TestTreeAppliedInSubtreeRelevant(t *testing.T) {
	attrs := example.Applied("subdir/a.dat")

	assert.Len(t, attrs, 5)
	assert.Equal(t, attrs[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, attrs[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, attrs[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, attrs[3], &Attr{K: "text", V: "false"})
	assert.Equal(t, attrs[4], &Attr{K: "subdir", V: "yes"})
}

func TestTreeAppliedInSubtreeIrrelevant(t *testing.T) {
	attrs := example.Applied("subdir/a.txt")

	assert.Empty(t, attrs)
}

func TestTreeAppliedInIrrelevantSubtree(t *testing.T) {
	attrs := example.Applied("other/subdir/a.dat")

	assert.Len(t, attrs, 4)
	assert.Equal(t, attrs[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, attrs[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, attrs[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, attrs[3], &Attr{K: "text", V: "false"})
}

func TestNewDiscoversSimpleTrees(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.Remove(tmp)

	db, err := gitobj.FromFilesystem(tmp, "")
	require.NoError(t, err)
	defer db.Close()

	blob, err := db.WriteBlob(gitobj.NewBlobFromBytes([]byte(`
		*.dat filter=lfs diff=lfs merge=lfs -text
	`)))
	require.NoError(t, err)

	tree, err := New(db, &gitobj.Tree{Entries: []*gitobj.TreeEntry{
		{
			Name:     ".gitattributes",
			Oid:      blob,
			Filemode: 0100644,
		},
	}})
	require.NoError(t, err)

	attrs := tree.Applied("foo.dat")

	assert.Len(t, attrs, 4)
	assert.Equal(t, attrs[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, attrs[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, attrs[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, attrs[3], &Attr{K: "text", V: "false"})
}

func TestNewDiscoversSimpleChildrenTrees(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.Remove(tmp)

	db, err := gitobj.FromFilesystem(tmp, "")
	require.NoError(t, err)
	defer db.Close()

	blob, err := db.WriteBlob(gitobj.NewBlobFromBytes([]byte(`
		*.dat filter=lfs diff=lfs merge=lfs -text
	`)))
	require.NoError(t, err)

	child, err := db.WriteTree(&gitobj.Tree{Entries: []*gitobj.TreeEntry{
		{
			Name:     ".gitattributes",
			Oid:      blob,
			Filemode: 0100644,
		},
	}})
	require.NoError(t, err)

	tree, err := New(db, &gitobj.Tree{Entries: []*gitobj.TreeEntry{
		{
			Name:     "child",
			Oid:      child,
			Filemode: 040000,
		},
	}})
	require.NoError(t, err)
	assert.Empty(t, tree.Lines)
	assert.Len(t, tree.Children, 1)

	attrs := tree.Applied("child/foo.dat")

	assert.Len(t, attrs, 4)
	assert.Equal(t, attrs[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, attrs[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, attrs[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, attrs[3], &Attr{K: "text", V: "false"})
}

func TestNewDiscoversIndirectChildrenTrees(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.Remove(tmp)

	db, err := gitobj.FromFilesystem(tmp, "")
	require.NoError(t, err)
	defer db.Close()

	blob, err := db.WriteBlob(gitobj.NewBlobFromBytes([]byte(`
		*.dat filter=lfs diff=lfs merge=lfs -text
	`)))
	require.NoError(t, err)

	indirect, err := db.WriteTree(&gitobj.Tree{Entries: []*gitobj.TreeEntry{
		{
			Name:     ".gitattributes",
			Oid:      blob,
			Filemode: 0100644,
		},
	}})
	require.NoError(t, err)

	child, err := db.WriteTree(&gitobj.Tree{Entries: []*gitobj.TreeEntry{
		{
			Name:     "indirect",
			Oid:      indirect,
			Filemode: 040000,
		},
	}})
	require.NoError(t, err)

	tree, err := New(db, &gitobj.Tree{Entries: []*gitobj.TreeEntry{
		{
			Name:     "child",
			Oid:      child,
			Filemode: 040000,
		},
	}})
	require.NoError(t, err)
	assert.Empty(t, tree.Lines)
	assert.Len(t, tree.Children, 1)

	attrs := tree.Applied("child/indirect/foo.dat")

	assert.Len(t, attrs, 4)
	assert.Equal(t, attrs[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, attrs[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, attrs[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, attrs[3], &Attr{K: "text", V: "false"})
}

func TestNewIgnoresChildrenAppropriately(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.Remove(tmp)

	db, err := gitobj.FromFilesystem(tmp, "")
	require.NoError(t, err)
	defer db.Close()

	blob, err := db.WriteBlob(gitobj.NewBlobFromBytes([]byte(`
		*.dat filter=lfs diff=lfs merge=lfs -text
	`)))
	require.NoError(t, err)

	child, err := db.WriteTree(&gitobj.Tree{Entries: []*gitobj.TreeEntry{
		{
			Name:     "README.md",
			Oid:      []byte("00000000000000000000"),
			Filemode: 0100644,
		},
	}})
	require.NoError(t, err)

	tree, err := New(db, &gitobj.Tree{Entries: []*gitobj.TreeEntry{
		{
			Name:     ".gitattributes",
			Oid:      blob,
			Filemode: 0100644,
		},
		{
			Name:     "child",
			Oid:      child,
			Filemode: 040000,
		},
	}})
	require.NoError(t, err)

	assert.NotContains(t, tree.Children, "child")
}
