package pack

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeIndexUnsupportedVersion(t *testing.T) {
	buf := make([]byte, 0, 4+4)
	buf = append(buf, 0xff, 0x74, 0x4f, 0x63)
	buf = append(buf, 0x0, 0x0, 0x0, 0x3)

	idx, err := DecodeIndex(bytes.NewReader(buf))

	assert.EqualError(t, err, "git/odb/pack: unsupported version: 3")
	assert.Nil(t, idx)
}

func TestDecodeIndexEmptyContents(t *testing.T) {
	idx, err := DecodeIndex(bytes.NewReader(make([]byte, 0)))

	assert.Equal(t, io.EOF, err)
	assert.Nil(t, idx)
}
