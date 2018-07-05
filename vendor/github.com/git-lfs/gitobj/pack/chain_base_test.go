package pack

import (
	"bytes"
	"compress/zlib"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChainBaseDecompressesData(t *testing.T) {
	const contents = "Hello, world!\n"

	compressed, err := compress(contents)
	assert.NoError(t, err)

	var buf bytes.Buffer

	_, err = buf.Write([]byte{0x0, 0x0, 0x0, 0x0})
	assert.NoError(t, err)

	_, err = buf.Write(compressed)
	assert.NoError(t, err)

	_, err = buf.Write([]byte{0x0, 0x0, 0x0, 0x0})
	assert.NoError(t, err)

	base := &ChainBase{
		offset: 4,
		size:   int64(len(contents)),

		r: bytes.NewReader(buf.Bytes()),
	}

	unpacked, err := base.Unpack()
	assert.NoError(t, err)
	assert.Equal(t, contents, string(unpacked))
}

func TestChainBaseTypeReturnsType(t *testing.T) {
	b := &ChainBase{
		typ: TypeCommit,
	}

	assert.Equal(t, TypeCommit, b.Type())
}

func compress(base string) ([]byte, error) {
	var buf bytes.Buffer

	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write([]byte(base)); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
