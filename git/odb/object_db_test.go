package odb

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeBlob(t *testing.T) {
	sha := "af5626b4a114abcb82d63db7c8082c3c4756e51b"
	contents := "Hello, world!\n"

	var buf bytes.Buffer

	zw := zlib.NewWriter(&buf)
	fmt.Fprintf(zw, "blob 14\x00%s", contents)
	zw.Close()

	odb := &ObjectDatabase{s: newMemoryStorer(map[string]io.ReadWriter{
		sha: &buf,
	})}

	shaHex, _ := hex.DecodeString(sha)
	blob, err := odb.Blob(shaHex)

	assert.Nil(t, err)
	assert.EqualValues(t, 14, blob.Size)

	got, err := ioutil.ReadAll(blob.Contents)
	assert.Nil(t, err)
	assert.Equal(t, contents, string(got))
}

func TestDecodeTree(t *testing.T) {
	sha := "fcb545d5746547a597811b7441ed8eba307be1ff"
	hexSha, err := hex.DecodeString(sha)
	require.Nil(t, err)

	blobSha := "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"
	hexBlobSha, err := hex.DecodeString(blobSha)
	require.Nil(t, err)

	var buf bytes.Buffer

	zw := zlib.NewWriter(&buf)
	fmt.Fprintf(zw, "tree 37\x00")
	fmt.Fprintf(zw, "100644 hello.txt\x00")
	zw.Write(hexBlobSha)
	zw.Close()

	odb := &ObjectDatabase{s: newMemoryStorer(map[string]io.ReadWriter{
		sha: &buf,
	})}

	tree, err := odb.Tree(hexSha)

	assert.Nil(t, err)
	require.Equal(t, 1, len(tree.Entries))
	assert.Equal(t, &TreeEntry{
		Name:     "hello.txt",
		Oid:      hexBlobSha,
		Filemode: 0100644,
	}, tree.Entries[0])
}

func TestDecodeCommit(t *testing.T) {
	sha := "d7283480bb6dc90be621252e1001a93871dcf511"
	commitShaHex, err := hex.DecodeString(sha)
	assert.Nil(t, err)

	var buf bytes.Buffer

	zw := zlib.NewWriter(&buf)
	fmt.Fprintf(zw, "commit 173\x00")
	fmt.Fprintf(zw, "tree fcb545d5746547a597811b7441ed8eba307be1ff\n")
	fmt.Fprintf(zw, "author Taylor Blau <me@ttaylorr.com> 1494620424 -0600\n")
	fmt.Fprintf(zw, "committer Taylor Blau <me@ttaylorr.com> 1494620424 -0600\n")
	fmt.Fprintf(zw, "\ninitial commit\n")
	zw.Close()

	odb := &ObjectDatabase{s: newMemoryStorer(map[string]io.ReadWriter{
		sha: &buf,
	})}

	commit, err := odb.Commit(commitShaHex)

	assert.Nil(t, err)
	assert.Equal(t, "Taylor Blau <me@ttaylorr.com> 1494620424 -0600", commit.Author)
	assert.Equal(t, "Taylor Blau <me@ttaylorr.com> 1494620424 -0600", commit.Committer)
	assert.Equal(t, "initial commit", commit.Message)
	assert.Equal(t, 0, len(commit.ParentIDs))
	assert.Equal(t, "fcb545d5746547a597811b7441ed8eba307be1ff", hex.EncodeToString(commit.TreeID))
}

func TestWriteBlob(t *testing.T) {
	fs := newMemoryStorer(make(map[string]io.ReadWriter))
	odb := &ObjectDatabase{s: fs}

	sha, err := odb.WriteBlob(&Blob{
		Size:     14,
		Contents: strings.NewReader("Hello, world!\n"),
	})

	expected := "af5626b4a114abcb82d63db7c8082c3c4756e51b"

	assert.Nil(t, err)
	assert.Equal(t, expected, hex.EncodeToString(sha))
	assert.NotNil(t, fs.fs[hex.EncodeToString(sha)])
}

func TestWriteTree(t *testing.T) {
	fs := newMemoryStorer(make(map[string]io.ReadWriter))
	odb := &ObjectDatabase{s: fs}

	blobSha := "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"
	hexBlobSha, err := hex.DecodeString(blobSha)
	require.Nil(t, err)

	sha, err := odb.WriteTree(&Tree{Entries: []*TreeEntry{
		{
			Name:     "hello.txt",
			Oid:      hexBlobSha,
			Filemode: 0100644,
		},
	}})

	expected := "fcb545d5746547a597811b7441ed8eba307be1ff"

	assert.Nil(t, err)
	assert.Equal(t, expected, hex.EncodeToString(sha))
	assert.NotNil(t, fs.fs[hex.EncodeToString(sha)])
}

func TestWriteCommit(t *testing.T) {
	fs := newMemoryStorer(make(map[string]io.ReadWriter))
	odb := &ObjectDatabase{s: fs}

	when := time.Unix(1257894000, 0).UTC()
	author := &Signature{Name: "John Doe", Email: "john@example.com", When: when}
	committer := &Signature{Name: "Jane Doe", Email: "jane@example.com", When: when}

	tree := "fcb545d5746547a597811b7441ed8eba307be1ff"
	treeHex, err := hex.DecodeString(tree)
	assert.Nil(t, err)

	sha, err := odb.WriteCommit(&Commit{
		Author:    author.String(),
		Committer: committer.String(),
		TreeID:    treeHex,
		Message:   "initial commit",
	})

	expected := "77a746376fdb591a44a4848b5ba308b2d3e2a90c"

	assert.Nil(t, err)
	assert.Equal(t, expected, hex.EncodeToString(sha))
	assert.NotNil(t, fs.fs[hex.EncodeToString(sha)])
}

func TestDecodeTag(t *testing.T) {
	const sha = "7639ba293cd2c457070e8446ecdea56682af0f48"
	tagShaHex, err := hex.DecodeString(sha)

	var buf bytes.Buffer

	zw := zlib.NewWriter(&buf)
	fmt.Fprintf(zw, "tag 165\x00")
	fmt.Fprintf(zw, "object 6161616161616161616161616161616161616161\n")
	fmt.Fprintf(zw, "type commit\n")
	fmt.Fprintf(zw, "tag v2.4.0\n")
	fmt.Fprintf(zw, "tagger A U Thor <author@example.com>\n")
	fmt.Fprintf(zw, "\n")
	fmt.Fprintf(zw, "The quick brown fox jumps over the lazy dog.\n")
	zw.Close()

	odb := &ObjectDatabase{s: newMemoryStorer(map[string]io.ReadWriter{
		sha: &buf,
	})}

	tag, err := odb.Tag(tagShaHex)

	assert.Nil(t, err)

	assert.Equal(t, []byte("aaaaaaaaaaaaaaaaaaaa"), tag.Object)
	assert.Equal(t, CommitObjectType, tag.ObjectType)
	assert.Equal(t, "v2.4.0", tag.Name)
	assert.Equal(t, "A U Thor <author@example.com>", tag.Tagger)
	assert.Equal(t, "The quick brown fox jumps over the lazy dog.", tag.Message)
}

func TestWriteTag(t *testing.T) {
	fs := newMemoryStorer(make(map[string]io.ReadWriter))
	odb := &ObjectDatabase{s: fs}

	sha, err := odb.WriteTag(&Tag{
		Object:     []byte("aaaaaaaaaaaaaaaaaaaa"),
		ObjectType: CommitObjectType,
		Name:       "v2.4.0",
		Tagger:     "A U Thor <author@example.com>",

		Message: "The quick brown fox jumps over the lazy dog.",
	})

	expected := "e614dda21829f4176d3db27fe62fb4aee2e2475d"

	assert.Nil(t, err)
	assert.Equal(t, expected, hex.EncodeToString(sha))
	assert.NotNil(t, fs.fs[hex.EncodeToString(sha)])
}

func TestReadingAMissingObjectAfterClose(t *testing.T) {
	sha, _ := hex.DecodeString("af5626b4a114abcb82d63db7c8082c3c4756e51b")

	db := &ObjectDatabase{
		s:      newMemoryStorer(nil),
		closed: 1,
	}

	blob, err := db.Blob(sha)
	assert.EqualError(t, err, "git/odb: cannot use closed *pack.Set")
	assert.Nil(t, blob)
}

func TestClosingAnObjectDatabaseMoreThanOnce(t *testing.T) {
	db, err := FromFilesystem("/tmp", "")
	assert.Nil(t, err)

	assert.Nil(t, db.Close())
	assert.EqualError(t, db.Close(), "git/odb: *ObjectDatabase already closed")
}

func TestObjectDatabaseRootWithRoot(t *testing.T) {
	db, err := FromFilesystem("/foo/bar/baz", "")
	assert.Nil(t, err)

	root, ok := db.Root()
	assert.Equal(t, "/foo/bar/baz", root)
	assert.True(t, ok)
}

func TestObjectDatabaseRootWithoutRoot(t *testing.T) {
	root, ok := new(ObjectDatabase).Root()

	assert.Equal(t, "", root)
	assert.False(t, ok)
}
