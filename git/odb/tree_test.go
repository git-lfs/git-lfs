package odb

import (
	"bufio"
	"bytes"
	"fmt"
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
