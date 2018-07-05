package pack

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOffsetReaderAtReadsAtOffset(t *testing.T) {
	bo := &OffsetReaderAt{
		r: bytes.NewReader([]byte{0x0, 0x1, 0x2, 0x3}),
		o: 1,
	}

	var x1 [1]byte
	n1, e1 := bo.Read(x1[:])

	assert.NoError(t, e1)
	assert.Equal(t, 1, n1)

	assert.EqualValues(t, 0x1, x1[0])

	var x2 [1]byte
	n2, e2 := bo.Read(x2[:])

	assert.NoError(t, e2)
	assert.Equal(t, 1, n2)
	assert.EqualValues(t, 0x2, x2[0])
}

func TestOffsetReaderPropogatesErrors(t *testing.T) {
	expected := fmt.Errorf("gitobj/pack: testing")
	bo := &OffsetReaderAt{
		r: &ErrReaderAt{Err: expected},
		o: 1,
	}

	n, err := bo.Read(make([]byte, 1))

	assert.Equal(t, expected, err)
	assert.Equal(t, 0, n)
}

type ErrReaderAt struct {
	Err error
}

func (e *ErrReaderAt) ReadAt(p []byte, at int64) (n int, err error) {
	return 0, e.Err
}
