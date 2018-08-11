package pack

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	V1IndexFanout = make([]uint32, indexFanoutEntries)

	V1IndexSmallEntry = []byte{
		0x0, 0x0, 0x0, 0x1,

		0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1,
		0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1,
	}
	V1IndexSmallSha = V1IndexSmallEntry[4:]

	V1IndexMediumEntry = []byte{
		0x0, 0x0, 0x0, 0x2,

		0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2,
		0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2,
	}
	V1IndexMediumSha = V1IndexMediumEntry[4:]

	V1IndexLargeEntry = []byte{
		0x0, 0x0, 0x0, 0x3,

		0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3,
		0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3,
	}
	V1IndexLargeSha = V1IndexLargeEntry[4:]

	V1Index = &Index{
		fanout:  V1IndexFanout,
		version: new(V1),
	}
)

func TestIndexV1SearchExact(t *testing.T) {
	e, err := new(V1).Entry(V1Index, 1)

	assert.NoError(t, err)
	assert.EqualValues(t, 2, e.PackOffset)
}

func TestIndexVersionWidthV1(t *testing.T) {
	assert.EqualValues(t, 0, new(V1).Width())
}

func init() {
	V1IndexFanout[1] = 1
	V1IndexFanout[2] = 2
	V1IndexFanout[3] = 3

	for i := 3; i < len(V1IndexFanout); i++ {
		V1IndexFanout[i] = 3
	}

	fanout := make([]byte, indexFanoutWidth)
	for i, n := range V1IndexFanout {
		binary.BigEndian.PutUint32(fanout[i*indexFanoutEntryWidth:], n)
	}

	buf := make([]byte, 0, indexOffsetV1Start+(3*indexObjectEntryV1Width))

	buf = append(buf, fanout...)
	buf = append(buf, V1IndexSmallEntry...)
	buf = append(buf, V1IndexMediumEntry...)
	buf = append(buf, V1IndexLargeEntry...)

	V1Index.r = bytes.NewReader(buf)
}
