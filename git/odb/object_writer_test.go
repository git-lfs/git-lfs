package odb

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectWriterWritesHeaders(t *testing.T) {
	var buf bytes.Buffer

	w := NewObjectWriter(&buf)

	n, err := w.WriteHeader(BlobObjectType, 1)
	assert.Equal(t, 7, n)
	assert.Nil(t, err)

	assert.Nil(t, w.Close())

	r, err := zlib.NewReader(&buf)
	assert.Nil(t, err)

	all, err := ioutil.ReadAll(r)
	assert.Nil(t, err)
	assert.Equal(t, []byte("blob 1\x00"), all)

	assert.Nil(t, r.Close())
}

func TestObjectWriterWritesData(t *testing.T) {
	var buf bytes.Buffer

	w := NewObjectWriter(&buf)
	w.WriteHeader(BlobObjectType, 1)

	n, err := w.Write([]byte{0x31})
	assert.Equal(t, 1, n)
	assert.Nil(t, err)

	assert.Nil(t, w.Close())

	r, err := zlib.NewReader(&buf)
	assert.Nil(t, err)

	all, err := ioutil.ReadAll(r)
	assert.Nil(t, err)
	assert.Equal(t, []byte("blob 1\x001"), all)

	assert.Nil(t, r.Close())
}

func TestObjectWriterPanicsOnWritesWithoutHeader(t *testing.T) {
	defer func() {
		err := recover()

		assert.NotNil(t, err)
		assert.Equal(t, "git/odb: cannot write data without header", err)
	}()

	w := NewObjectWriter(new(bytes.Buffer))
	w.Write(nil)
}

func TestObjectWriterPanicsOnMultipleHeaderWrites(t *testing.T) {
	defer func() {
		err := recover()

		assert.NotNil(t, err)
		assert.Equal(t, "git/odb: cannot write headers more than once", err)
	}()

	w := NewObjectWriter(new(bytes.Buffer))
	w.WriteHeader(BlobObjectType, 1)
	w.WriteHeader(TreeObjectType, 2)
}

func TestObjectWriterKeepsTrackOfHash(t *testing.T) {
	w := NewObjectWriter(new(bytes.Buffer))
	n, err := w.WriteHeader(BlobObjectType, 1)

	assert.Nil(t, err)
	assert.Equal(t, 7, n)

	assert.Equal(t, "bb6ca78b66403a67c6281df142de5ef472186283", hex.EncodeToString(w.Sha()))
}

type WriteCloserFn struct {
	io.Writer
	closeFn func() error
}

func (r *WriteCloserFn) Close() error { return r.closeFn() }

func TestObjectWriterCallsClose(t *testing.T) {
	var calls uint32

	expected := errors.New("close error")

	w := NewObjectWriteCloser(&WriteCloserFn{
		Writer: new(bytes.Buffer),
		closeFn: func() error {
			atomic.AddUint32(&calls, 1)
			return expected
		},
	})

	got := w.Close()

	assert.EqualValues(t, 1, calls)
	assert.Equal(t, expected, got)
}
