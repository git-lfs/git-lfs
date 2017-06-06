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
				Type:     BlobObjectType,
				Oid:      []byte("aaaaaaaaaaaaaaaaaaaa"),
				Filemode: 0100644,
			},
			{
				Name:     "subdir",
				Type:     TreeObjectType,
				Oid:      []byte("bbbbbbbbbbbbbbbbbbbb"),
				Filemode: 040000,
			},
			{
				Name:     "submodule",
				Type:     CommitObjectType,
				Oid:      []byte("cccccccccccccccccccc"),
				Filemode: 0160000,
			},
		},
	}

	buf := new(bytes.Buffer)

	n, err := tree.Encode(buf)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, n)

	assertTreeEntry(t, buf, "a.dat", BlobObjectType, []byte("aaaaaaaaaaaaaaaaaaaa"), 0100644)
	assertTreeEntry(t, buf, "subdir", TreeObjectType, []byte("bbbbbbbbbbbbbbbbbbbb"), 040000)
	assertTreeEntry(t, buf, "submodule", CommitObjectType, []byte("cccccccccccccccccccc"), 0160000)

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
		Type:     BlobObjectType,
		Oid:      []byte("aaaaaaaaaaaaaaaaaaaa"),
		Filemode: 0100644,
	}, tree.Entries[0])
	assert.Equal(t, &TreeEntry{
		Name:     "subdir",
		Type:     TreeObjectType,
		Oid:      []byte("bbbbbbbbbbbbbbbbbbbb"),
		Filemode: 040000,
	}, tree.Entries[1])
	assert.Equal(t, &TreeEntry{
		Name:     "symlink",
		Type:     BlobObjectType,
		Oid:      []byte("cccccccccccccccccccc"),
		Filemode: 0120000,
	}, tree.Entries[2])
	assert.Equal(t, &TreeEntry{
		Name:     "submodule",
		Type:     CommitObjectType,
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
		Type:     BlobObjectType,
		Oid:      []byte("aaaaaaaaaaaaaaaaaaaa"),
		Filemode: 0100644,
	}, tree.Entries[0])
}

func assertTreeEntry(t *testing.T, buf *bytes.Buffer,
	name string, typ ObjectType, oid []byte, mode int32) {

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
