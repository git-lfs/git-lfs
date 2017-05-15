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

	"github.com/stretchr/testify/assert"
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
