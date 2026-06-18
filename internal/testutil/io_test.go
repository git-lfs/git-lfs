package testutil

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEagerEOFByteReaderReturnsEOFWithFinalByteRead(t *testing.T) {
	buf := []byte{0x1}
	bufSize := len(buf)

	r := NewEagerEOFByteReader(buf)

	p := make([]byte, bufSize+1)

	n, err := r.Read(p)

	assert.Equal(t, bufSize, n)
	assert.Equal(t, io.EOF, err)
}

func TestDeferredEOFByteReaderReturnsEOFAfterFinalByteRead(t *testing.T) {
	buf := []byte{0x1}
	bufSize := len(buf)

	r := NewDeferredEOFByteReader(buf)

	p := make([]byte, bufSize+1)

	n, err := r.Read(p)

	assert.Equal(t, bufSize, n)
	assert.Nil(t, err)

	n, err = r.Read(p)
	assert.Zero(t, n)
	assert.Equal(t, io.EOF, err)
}
