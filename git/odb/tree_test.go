package odb

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTreeReturnsCorrectObjectType(t *testing.T) {
	assert.Equal(t, TreeObjectType, new(Tree).Type())
}

func TestTreeEncoding(t *testing.T) {
	tree := &Tree{
		Entries: []*TreeEntry{
			{
				Name:     "a.dat",
				Oid:      []byte("aaaaaaaaaaaaaaaaaaaa"),
				Filemode: 0100644,
			},
			{
				Name:     "subdir",
				Oid:      []byte("bbbbbbbbbbbbbbbbbbbb"),
				Filemode: 040000,
			},
			{
				Name:     "submodule",
				Oid:      []byte("cccccccccccccccccccc"),
				Filemode: 0160000,
			},
		},
	}

	buf := new(bytes.Buffer)

	n, err := tree.Encode(buf)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, n)

	assertTreeEntry(t, buf, "a.dat", []byte("aaaaaaaaaaaaaaaaaaaa"), 0100644)
	assertTreeEntry(t, buf, "subdir", []byte("bbbbbbbbbbbbbbbbbbbb"), 040000)
	assertTreeEntry(t, buf, "submodule", []byte("cccccccccccccccccccc"), 0160000)

	assert.Equal(t, 0, buf.Len())
}

func TestTreeDecoding(t *testing.T) {
	from := new(bytes.Buffer)
	fmt.Fprintf(from, "%s %s\x00%s",
		strconv.FormatInt(int64(0100644), 8),
		"a.dat", []byte("aaaaaaaaaaaaaaaaaaaa"))
	fmt.Fprintf(from, "%s %s\x00%s",
		strconv.FormatInt(int64(040000), 8),
		"subdir", []byte("bbbbbbbbbbbbbbbbbbbb"))
	fmt.Fprintf(from, "%s %s\x00%s",
		strconv.FormatInt(int64(0120000), 8),
		"symlink", []byte("cccccccccccccccccccc"))
	fmt.Fprintf(from, "%s %s\x00%s",
		strconv.FormatInt(int64(0160000), 8),
		"submodule", []byte("dddddddddddddddddddd"))

	flen := from.Len()

	tree := new(Tree)
	n, err := tree.Decode(from, int64(flen))

	assert.Nil(t, err)
	assert.Equal(t, flen, n)

	require.Equal(t, 4, len(tree.Entries))
	assert.Equal(t, &TreeEntry{
		Name:     "a.dat",
		Oid:      []byte("aaaaaaaaaaaaaaaaaaaa"),
		Filemode: 0100644,
	}, tree.Entries[0])
	assert.Equal(t, &TreeEntry{
		Name:     "subdir",
		Oid:      []byte("bbbbbbbbbbbbbbbbbbbb"),
		Filemode: 040000,
	}, tree.Entries[1])
	assert.Equal(t, &TreeEntry{
		Name:     "symlink",
		Oid:      []byte("cccccccccccccccccccc"),
		Filemode: 0120000,
	}, tree.Entries[2])
	assert.Equal(t, &TreeEntry{
		Name:     "submodule",
		Oid:      []byte("dddddddddddddddddddd"),
		Filemode: 0160000,
	}, tree.Entries[3])
}

func TestTreeDecodingShaBoundary(t *testing.T) {
	var from bytes.Buffer

	fmt.Fprintf(&from, "%s %s\x00%s",
		strconv.FormatInt(int64(0100644), 8),
		"a.dat", []byte("aaaaaaaaaaaaaaaaaaaa"))

	flen := from.Len()

	tree := new(Tree)
	n, err := tree.Decode(bufio.NewReaderSize(&from, flen-2), int64(flen))

	assert.Nil(t, err)
	assert.Equal(t, flen, n)

	require.Len(t, tree.Entries, 1)
	assert.Equal(t, &TreeEntry{
		Name:     "a.dat",
		Oid:      []byte("aaaaaaaaaaaaaaaaaaaa"),
		Filemode: 0100644,
	}, tree.Entries[0])
}

func TestTreeMergeReplaceElements(t *testing.T) {
	e1 := &TreeEntry{Name: "a", Filemode: 0100644, Oid: []byte{0x1}}
	e2 := &TreeEntry{Name: "b", Filemode: 0100644, Oid: []byte{0x2}}
	e3 := &TreeEntry{Name: "c", Filemode: 0100644, Oid: []byte{0x3}}

	e4 := &TreeEntry{Name: "b", Filemode: 0100644, Oid: []byte{0x4}}
	e5 := &TreeEntry{Name: "c", Filemode: 0100644, Oid: []byte{0x5}}

	t1 := &Tree{Entries: []*TreeEntry{e1, e2, e3}}

	t2 := t1.Merge(e4, e5)

	require.Len(t, t1.Entries, 3)
	assert.True(t, bytes.Equal(t1.Entries[0].Oid, []byte{0x1}))
	assert.True(t, bytes.Equal(t1.Entries[1].Oid, []byte{0x2}))
	assert.True(t, bytes.Equal(t1.Entries[2].Oid, []byte{0x3}))

	require.Len(t, t2.Entries, 3)
	assert.True(t, bytes.Equal(t2.Entries[0].Oid, []byte{0x1}))
	assert.True(t, bytes.Equal(t2.Entries[1].Oid, []byte{0x4}))
	assert.True(t, bytes.Equal(t2.Entries[2].Oid, []byte{0x5}))
}

func TestMergeInsertElementsInSubtreeOrder(t *testing.T) {
	e1 := &TreeEntry{Name: "a-b", Filemode: 0100644, Oid: []byte{0x1}}
	e2 := &TreeEntry{Name: "a", Filemode: 040000, Oid: []byte{0x2}}
	e3 := &TreeEntry{Name: "a=", Filemode: 0100644, Oid: []byte{0x3}}
	e4 := &TreeEntry{Name: "a-", Filemode: 0100644, Oid: []byte{0x4}}

	t1 := &Tree{Entries: []*TreeEntry{e1, e2, e3}}
	t2 := t1.Merge(e4)

	require.Len(t, t1.Entries, 3)
	assert.True(t, bytes.Equal(t1.Entries[0].Oid, []byte{0x1}))
	assert.True(t, bytes.Equal(t1.Entries[1].Oid, []byte{0x2}))
	assert.True(t, bytes.Equal(t1.Entries[2].Oid, []byte{0x3}))

	assert.True(t, bytes.Equal(t2.Entries[0].Oid, []byte{0x4}))
	assert.True(t, bytes.Equal(t2.Entries[1].Oid, []byte{0x1}))
	assert.True(t, bytes.Equal(t2.Entries[2].Oid, []byte{0x2}))
	assert.True(t, bytes.Equal(t2.Entries[3].Oid, []byte{0x3}))
}

type TreeEntryTypeTestCase struct {
	Filemode int32
	Expected ObjectType
}

func (c *TreeEntryTypeTestCase) Assert(t *testing.T) {
	e := &TreeEntry{Filemode: c.Filemode}

	got := e.Type()

	assert.Equal(t, c.Expected, got,
		"git/odb: expected type: %s, got: %s", c.Expected, got)
}

func TestTreeEntryTypeResolution(t *testing.T) {
	for desc, c := range map[string]*TreeEntryTypeTestCase{
		"blob":    {0100644, BlobObjectType},
		"subtree": {040000, TreeObjectType},
		"symlink": {0120000, BlobObjectType},
		"commit":  {0160000, CommitObjectType},
	} {
		t.Run(desc, c.Assert)
	}
}

func TestTreeEntryTypeResolutionUnknown(t *testing.T) {
	e := &TreeEntry{Filemode: -1}

	defer func() {
		if err := recover(); err == nil {
			t.Fatal("git/odb: expected panic(), got none")
		} else {
			assert.Equal(t, "git/odb: unknown object type: -1", err)
		}
	}()

	e.Type()
}

func TestSubtreeOrder(t *testing.T) {
	// The below list (e1, e2, ..., e5) is entered in subtree order: that
	// is, lexicographically byte-ordered as if blobs end in a '\0', and
	// sub-trees end in a '/'.
	//
	// See:
	//   http://public-inbox.org/git/7vac6jfzem.fsf@assigned-by-dhcp.cox.net
	e1 := &TreeEntry{Filemode: 0100644, Name: "a-"}
	e2 := &TreeEntry{Filemode: 0100644, Name: "a-b"}
	e3 := &TreeEntry{Filemode: 040000, Name: "a"}
	e4 := &TreeEntry{Filemode: 0100644, Name: "a="}
	e5 := &TreeEntry{Filemode: 0100644, Name: "a=b"}

	// Create a set of entries in the wrong order:
	entries := []*TreeEntry{e3, e4, e1, e5, e2}

	sort.Sort(SubtreeOrder(entries))

	// Assert that they are in the correct order after sorting in sub-tree
	// order:
	require.Len(t, entries, 5)
	assert.Equal(t, "a-", entries[0].Name)
	assert.Equal(t, "a-b", entries[1].Name)
	assert.Equal(t, "a", entries[2].Name)
	assert.Equal(t, "a=", entries[3].Name)
	assert.Equal(t, "a=b", entries[4].Name)
}

func TestSubtreeOrderReturnsEmptyForOutOfBounds(t *testing.T) {
	o := SubtreeOrder([]*TreeEntry{{Name: "a"}})

	assert.Equal(t, "", o.Name(len(o)+1))
}

func TestSubtreeOrderReturnsEmptyForNilElements(t *testing.T) {
	o := SubtreeOrder([]*TreeEntry{nil})

	assert.Equal(t, "", o.Name(0))
}

func TestTreeEqualReturnsTrueWithUnchangedContents(t *testing.T) {
	t1 := &Tree{Entries: []*TreeEntry{
		{Name: "a.dat", Filemode: 0100644, Oid: make([]byte, 20)},
	}}
	t2 := &Tree{Entries: []*TreeEntry{
		{Name: "a.dat", Filemode: 0100644, Oid: make([]byte, 20)},
	}}

	assert.True(t, t1.Equal(t2))
}

func TestTreeEqualReturnsFalseWithChangedContents(t *testing.T) {
	t1 := &Tree{Entries: []*TreeEntry{
		{Name: "a.dat", Filemode: 0100644, Oid: make([]byte, 20)},
		{Name: "b.dat", Filemode: 0100644, Oid: make([]byte, 20)},
	}}
	t2 := &Tree{Entries: []*TreeEntry{
		{Name: "a.dat", Filemode: 0100644, Oid: make([]byte, 20)},
		{Name: "c.dat", Filemode: 0100644, Oid: make([]byte, 20)},
	}}

	assert.False(t, t1.Equal(t2))
}

func TestTreeEqualReturnsTrueWhenOneTreeIsNil(t *testing.T) {
	t1 := &Tree{Entries: []*TreeEntry{
		{Name: "a.dat", Filemode: 0100644, Oid: make([]byte, 20)},
	}}
	t2 := (*Tree)(nil)

	assert.False(t, t1.Equal(t2))
	assert.False(t, t2.Equal(t1))
}

func TestTreeEqualReturnsTrueWhenBothTreesAreNil(t *testing.T) {
	t1 := (*Tree)(nil)
	t2 := (*Tree)(nil)

	assert.True(t, t1.Equal(t2))
}

func TestTreeEntryEqualReturnsTrueWhenEntriesAreTheSame(t *testing.T) {
	e1 := &TreeEntry{Name: "a.dat", Filemode: 0100644, Oid: make([]byte, 20)}
	e2 := &TreeEntry{Name: "a.dat", Filemode: 0100644, Oid: make([]byte, 20)}

	assert.True(t, e1.Equal(e2))
}

func TestTreeEntryEqualReturnsFalseWhenDifferentNames(t *testing.T) {
	e1 := &TreeEntry{Name: "a.dat", Filemode: 0100644, Oid: make([]byte, 20)}
	e2 := &TreeEntry{Name: "b.dat", Filemode: 0100644, Oid: make([]byte, 20)}

	assert.False(t, e1.Equal(e2))
}

func TestTreeEntryEqualReturnsFalseWhenDifferentOids(t *testing.T) {
	e1 := &TreeEntry{Name: "a.dat", Filemode: 0100644, Oid: make([]byte, 20)}
	e2 := &TreeEntry{Name: "a.dat", Filemode: 0100644, Oid: make([]byte, 20)}

	e2.Oid[0] = 1

	assert.False(t, e1.Equal(e2))
}

func TestTreeEntryEqualReturnsFalseWhenDifferentFilemodes(t *testing.T) {
	e1 := &TreeEntry{Name: "a.dat", Filemode: 0100644, Oid: make([]byte, 20)}
	e2 := &TreeEntry{Name: "a.dat", Filemode: 0100755, Oid: make([]byte, 20)}

	assert.False(t, e1.Equal(e2))
}

func TestTreeEntryEqualReturnsFalseWhenOneEntryIsNil(t *testing.T) {
	e1 := &TreeEntry{Name: "a.dat", Filemode: 0100644, Oid: make([]byte, 20)}
	e2 := (*TreeEntry)(nil)

	assert.False(t, e1.Equal(e2))
}

func TestTreeEntryEqualReturnsTrueWhenBothEntriesAreNil(t *testing.T) {
	e1 := (*TreeEntry)(nil)
	e2 := (*TreeEntry)(nil)

	assert.True(t, e1.Equal(e2))
}

func assertTreeEntry(t *testing.T, buf *bytes.Buffer,
	name string, oid []byte, mode int32) {

	fmode, err := buf.ReadBytes(' ')
	assert.Nil(t, err)
	assert.Equal(t, []byte(strconv.FormatInt(int64(mode), 8)+" "), fmode)

	fname, err := buf.ReadBytes('\x00')
	assert.Nil(t, err)
	assert.Equal(t, []byte(name+"\x00"), fname)

	var sha [20]byte
	_, err = buf.Read(sha[:])
	assert.Nil(t, err)
	assert.Equal(t, oid, sha[:])
}
