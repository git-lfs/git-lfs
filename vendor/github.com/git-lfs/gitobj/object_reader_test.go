package gitobj

import (
	"bytes"
	"compress/zlib"
	"errors"
	"io"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectReaderReadsHeaders(t *testing.T) {
	var compressed bytes.Buffer

	zw := zlib.NewWriter(&compressed)
	zw.Write([]byte("blob 1\x00"))
	zw.Close()

	or, err := NewObjectReader(&compressed)
	assert.Nil(t, err)

	typ, size, err := or.Header()

	assert.Nil(t, err)
	assert.EqualValues(t, 1, size)
	assert.Equal(t, BlobObjectType, typ)
}

func TestObjectReaderConsumesHeaderBeforeReads(t *testing.T) {
	var compressed bytes.Buffer

	zw := zlib.NewWriter(&compressed)
	zw.Write([]byte("blob 1\x00asdf"))
	zw.Close()

	or, err := NewObjectReader(&compressed)
	assert.Nil(t, err)

	var buf [4]byte
	n, err := or.Read(buf[:])

	assert.Equal(t, 4, n)
	assert.Equal(t, []byte{'a', 's', 'd', 'f'}, buf[:])
	assert.Nil(t, err)
}

type ReadCloserFn struct {
	io.Reader
	closeFn func() error
}

func (r *ReadCloserFn) Close() error {
	return r.closeFn()
}

func TestObjectReaderCallsClose(t *testing.T) {
	var calls uint32
	expected := errors.New("expected")

	or, err := NewObjectReadCloser(&ReadCloserFn{
		Reader: bytes.NewBuffer([]byte{0x78, 0x01}),
		closeFn: func() error {
			atomic.AddUint32(&calls, 1)
			return expected
		},
	})
	assert.Nil(t, err)

	got := or.Close()

	assert.Equal(t, expected, got)
	assert.EqualValues(t, 1, atomic.LoadUint32(&calls))

}
