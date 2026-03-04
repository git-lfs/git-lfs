package gitattr

import (
	"testing"

	"github.com/git-lfs/gitobj/v2"
	"github.com/git-lfs/wildmatch/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	dat = wildmatch.NewWildmatch("*.dat",
		wildmatch.Basename,
		wildmatch.SystemCase)
	mp      = NewMacroProcessor()
	example = &Tree{
		mp: mp,
		lines: []Line{
			&patternLine{
				pattern: dat,
				lineAttrs: lineAttrs{
					attrs: []*Attr{
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
				},
			},
		},
		children: map[string]*Tree{
			"subdir": {
				mp: mp,
				lines: []Line{
					&patternLine{
						pattern: dat,
						lineAttrs: lineAttrs{
							attrs: []*Attr{
								{
									K: "subdir", V: "yes",
								},
							},
						},
					},
				},
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
	tmp := t.TempDir()

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
	tmp := t.TempDir()

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
	assert.Empty(t, tree.lines)
	assert.Len(t, tree.children, 1)

	attrs := tree.Applied("child/foo.dat")

	assert.Len(t, attrs, 4)
	assert.Equal(t, attrs[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, attrs[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, attrs[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, attrs[3], &Attr{K: "text", V: "false"})
}

func TestNewDiscoversIndirectChildrenTrees(t *testing.T) {
	tmp := t.TempDir()

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
	assert.Empty(t, tree.lines)
	assert.Len(t, tree.children, 1)

	attrs := tree.Applied("child/indirect/foo.dat")

	assert.Len(t, attrs, 4)
	assert.Equal(t, attrs[0], &Attr{K: "filter", V: "lfs"})
	assert.Equal(t, attrs[1], &Attr{K: "diff", V: "lfs"})
	assert.Equal(t, attrs[2], &Attr{K: "merge", V: "lfs"})
	assert.Equal(t, attrs[3], &Attr{K: "text", V: "false"})
}

func TestNewIgnoresChildrenAppropriately(t *testing.T) {
	tmp := t.TempDir()

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

	assert.NotContains(t, tree.children, "child")
}

func TestNewDiscoversSimpleTreesMacro(t *testing.T) {
	tmp := t.TempDir()

	db, err := gitobj.FromFilesystem(tmp, "")
	require.NoError(t, err)
	defer db.Close()

	blob, err := db.WriteBlob(gitobj.NewBlobFromBytes([]byte(`
	    [attr]lfs filter=lfs diff=lfs merge=lfs -text
		*.dat lfs
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

	assert.Len(t, attrs, 5)
	assert.Equal(t, &Attr{K: "filter", V: "lfs"}, attrs[0])
	assert.Equal(t, &Attr{K: "diff", V: "lfs"}, attrs[1])
	assert.Equal(t, &Attr{K: "merge", V: "lfs"}, attrs[2])
	assert.Equal(t, &Attr{K: "text", V: "false"}, attrs[3])
	assert.Equal(t, &Attr{K: "lfs", V: "true"}, attrs[4])
}

func TestAppliedProcessInCorrectOrder(t *testing.T) {
	tmp := t.TempDir()

	db, err := gitobj.FromFilesystem(tmp, "")
	require.NoError(t, err)
	defer db.Close()

	blob, err := db.WriteBlob(gitobj.NewBlobFromBytes([]byte(`
	    [attr]lfs filter=lfs diff=lfs merge=lfs -text
		*.dat lfs
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

	sysTree := &Tree{
		mp: tree.mp,
		lines: []Line{
			&patternLine{
				pattern: wildmatch.NewWildmatch("*.bin"),
				lineAttrs: lineAttrs{
					attrs: []*Attr{
						{K: "binary", V: "true"},
					},
				},
			},
		},
	}

	userTree := &Tree{
		mp: tree.mp,
		lines: []Line{
			&patternLine{
				pattern: wildmatch.NewWildmatch("*.png"),
				lineAttrs: lineAttrs{
					attrs: []*Attr{
						{K: "lfs", V: "true"},
					},
				},
			},
		},
	}

	repoTree := &Tree{
		mp: tree.mp,
		lines: []Line{
			&patternLine{
				pattern: wildmatch.NewWildmatch("*.dat"),
				lineAttrs: lineAttrs{
					attrs: []*Attr{
						{K: "diff", Unspecified: true},
					},
				},
			},
		},
	}

	tree.systemAttributes = sysTree
	tree.userAttributes = userTree
	tree.repoAttributes = repoTree

	attrs := tree.Applied("foo.dat")

	assert.Len(t, attrs, 6)
	assert.Equal(t, &Attr{K: "filter", V: "lfs"}, attrs[0])
	assert.Equal(t, &Attr{K: "diff", V: "lfs"}, attrs[1])
	assert.Equal(t, &Attr{K: "merge", V: "lfs"}, attrs[2])
	assert.Equal(t, &Attr{K: "text", V: "false"}, attrs[3])
	assert.Equal(t, &Attr{K: "lfs", V: "true"}, attrs[4])
	assert.Equal(t, &Attr{K: "diff", Unspecified: true}, attrs[5])

	attrs = tree.Applied("foo.bin")

	assert.Len(t, attrs, 4)
	assert.Equal(t, &Attr{K: "diff", V: "false"}, attrs[0])
	assert.Equal(t, &Attr{K: "merge", V: "false"}, attrs[1])
	assert.Equal(t, &Attr{K: "text", V: "false"}, attrs[2])
	assert.Equal(t, &Attr{K: "binary", V: "true"}, attrs[3])

	attrs = tree.Applied("foo.png")

	assert.Len(t, attrs, 5)
	assert.Equal(t, &Attr{K: "filter", V: "lfs"}, attrs[0])
	assert.Equal(t, &Attr{K: "diff", V: "lfs"}, attrs[1])
	assert.Equal(t, &Attr{K: "merge", V: "lfs"}, attrs[2])
	assert.Equal(t, &Attr{K: "text", V: "false"}, attrs[3])
	assert.Equal(t, &Attr{K: "lfs", V: "true"}, attrs[4])
}
