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

	odb := &ObjectDatabase{s: NewMemoryStorer(map[string]io.ReadWriter{
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

	odb := &ObjectDatabase{s: NewMemoryStorer(map[string]io.ReadWriter{
		sha: &buf,
	})}

	tree, err := odb.Tree(hexSha)

	assert.Nil(t, err)
	require.Equal(t, 1, len(tree.Entries))
	assert.Equal(t, &TreeEntry{
		Name:     "hello.txt",
		Type:     BlobObjectType,
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

	odb := &ObjectDatabase{s: NewMemoryStorer(map[string]io.ReadWriter{
		sha: &buf,
	})}

	commit, err := odb.Commit(commitShaHex)

	assert.Nil(t, err)
	assert.Equal(t, "Taylor Blau", commit.Author.Name)
	assert.Equal(t, "me@ttaylorr.com", commit.Author.Email)
	assert.Equal(t, "Taylor Blau", commit.Committer.Name)
	assert.Equal(t, "me@ttaylorr.com", commit.Committer.Email)
	assert.Equal(t, "initial commit", commit.Message)
	assert.Equal(t, 0, len(commit.ParentIds))
	assert.Equal(t, "fcb545d5746547a597811b7441ed8eba307be1ff", hex.EncodeToString(commit.TreeId))
}

func TestWriteBlob(t *testing.T) {
	fs := NewMemoryStorer(make(map[string]io.ReadWriter))
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
	fs := NewMemoryStorer(make(map[string]io.ReadWriter))
	odb := &ObjectDatabase{s: fs}

	blobSha := "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"
	hexBlobSha, err := hex.DecodeString(blobSha)
	require.Nil(t, err)

	sha, err := odb.WriteTree(&Tree{Entries: []*TreeEntry{
		{
			Name:     "hello.txt",
			Type:     BlobObjectType,
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
	fs := NewMemoryStorer(make(map[string]io.ReadWriter))
	odb := &ObjectDatabase{s: fs}

	when := time.Unix(1494258422, 0)
	author := &Signature{Name: "John Doe", Email: "john@example.com", When: when}
	committer := &Signature{Name: "Jane Doe", Email: "jane@example.com", When: when}

	tree := "fcb545d5746547a597811b7441ed8eba307be1ff"
	treeHex, err := hex.DecodeString(tree)
	assert.Nil(t, err)

	sha, err := odb.WriteCommit(&Commit{
		Author:    author,
		Committer: committer,
		TreeId:    treeHex,
		Message:   "initial commit",
	})

	expected := "5ba51ea684770de6fef57a94734c4f4a02cb361b"

	assert.Nil(t, err)
	assert.Equal(t, expected, hex.EncodeToString(sha))
	assert.NotNil(t, fs.fs[hex.EncodeToString(sha)])
}
