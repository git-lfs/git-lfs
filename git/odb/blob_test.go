package odb

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlobReturnsCorrectObjectType(t *testing.T) {
	assert.Equal(t, BlobObjectType, new(Blob).Type())
}

func TestBlobFromString(t *testing.T) {
	given := []byte("example")
	glen := len(given)

	b := NewBlobFromBytes(given)

	assert.EqualValues(t, glen, b.Size)

	contents, err := ioutil.ReadAll(b.Contents)
	assert.NoError(t, err)
	assert.Equal(t, given, contents)
}

func TestBlobEncoding(t *testing.T) {
	const contents = "Hello, world!\n"

	b := &Blob{
		Size:     int64(len(contents)),
		Contents: strings.NewReader(contents),
	}

	var buf bytes.Buffer
	if _, err := b.Encode(&buf); err != nil {
		t.Fatal(err.Error())
	}
	assert.Equal(t, contents, (&buf).String())
}

func TestBlobDecoding(t *testing.T) {
	const contents = "Hello, world!\n"
	from := strings.NewReader(contents)

	b := new(Blob)
	n, err := b.Decode(from, int64(len(contents)))

	assert.Equal(t, 0, n)
	assert.Nil(t, err)

	assert.EqualValues(t, len(contents), b.Size)

	got, err := ioutil.ReadAll(b.Contents)
	assert.Nil(t, err)
	assert.Equal(t, []byte(contents), got)
}

func TestBlobCallCloseFn(t *testing.T) {
	var calls uint32

	expected := errors.New("some close error")

	b := &Blob{
		closeFn: func() error {
			atomic.AddUint32(&calls, 1)
			return expected
		},
	}

	got := b.Close()

	assert.Equal(t, expected, got)
	assert.EqualValues(t, 1, calls)
}

func TestBlobCanCloseWithoutCloseFn(t *testing.T) {
	b := &Blob{
		closeFn: nil,
	}

	assert.Nil(t, b.Close())
}
