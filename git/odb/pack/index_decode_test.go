package pack

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeIndexV1InvalidFanout(t *testing.T) {
	idx, err := DecodeIndex(bytes.NewReader(make([]byte, indexFanoutWidth-1)))

	assert.Equal(t, ErrShortFanout, err)
	assert.Nil(t, idx)
}

func TestDecodeIndexV2(t *testing.T) {
	buf := make([]byte, 0, indexV2Width+indexFanoutWidth)
	buf = append(buf, 0xff, 0x74, 0x4f, 0x63)
	buf = append(buf, 0x0, 0x0, 0x0, 0x2)
	for i := 0; i < indexFanoutEntries; i++ {
		x := make([]byte, 4)

		binary.BigEndian.PutUint32(x, uint32(3))

		buf = append(buf, x...)
	}

	idx, err := DecodeIndex(bytes.NewReader(buf))

	assert.NoError(t, err)
	assert.EqualValues(t, 3, idx.Count())
}

func TestDecodeIndexV2InvalidFanout(t *testing.T) {
	buf := make([]byte, 0, indexV2Width+indexFanoutWidth-indexFanoutEntryWidth)
	buf = append(buf, 0xff, 0x74, 0x4f, 0x63)
	buf = append(buf, 0x0, 0x0, 0x0, 0x2)
	buf = append(buf, make([]byte, indexFanoutWidth-1)...)

	idx, err := DecodeIndex(bytes.NewReader(buf))

	assert.Nil(t, idx)
	assert.Equal(t, ErrShortFanout, err)
}

func TestDecodeIndexV1(t *testing.T) {
	idx, err := DecodeIndex(bytes.NewReader(make([]byte, indexFanoutWidth)))

	assert.NoError(t, err)
	assert.EqualValues(t, 0, idx.Count())
}

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
