package pack

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	V2IndexHeader = []byte{
		0xff, 0x74, 0x4f, 0x63,
		0x00, 0x00, 0x00, 0x02,
	}
	V2IndexFanout = make([]uint32, FanoutEntries)

	V2IndexNames = []byte{
		0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1,
		0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1,

		0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2,
		0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2,

		0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3,
		0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3, 0x3,
	}
	V2IndexSmallSha  = V2IndexNames[0:20]
	V2IndexMediumSha = V2IndexNames[20:40]
	V2IndexLargeSha  = V2IndexNames[40:60]

	V2IndexCRCs = []byte{
		0x0, 0x0, 0x0, 0x0,
		0x1, 0x1, 0x1, 0x1,
		0x2, 0x2, 0x2, 0x2,
	}

	V2IndexOffsets = []byte{
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x02,
		0x80, 0x00, 0x04, 0x5c,

		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03,
	}

	V2Index = &Index{
		fanout:  V2IndexFanout,
		version: V2,
	}
)

func TestIndexV2SearchExact(t *testing.T) {
	e, cmp, err := V2.Search(V2Index, V2IndexMediumSha, 1)

	assert.Equal(t, 0, cmp)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, e.PackOffset)
}

func TestIndexV2SearchSmall(t *testing.T) {
	e, cmp, err := V2.Search(V2Index, V2IndexMediumSha, 0)

	assert.Equal(t, 1, cmp)
	assert.NoError(t, err)
	assert.Nil(t, e)
}

func TestIndexV2SearchBig(t *testing.T) {
	e, cmp, err := V2.Search(V2Index, V2IndexMediumSha, 2)

	assert.Equal(t, -1, cmp)
	assert.NoError(t, err)
	assert.Nil(t, e)
}

func TestIndexV2SearchExtendedOffset(t *testing.T) {
	e, cmp, err := V2.Search(V2Index, V2IndexLargeSha, 2)

	assert.Equal(t, 0, cmp)
	assert.NoError(t, err)
	assert.EqualValues(t, 3, e.PackOffset)
}

func init() {
	V2IndexFanout[1] = 1
	V2IndexFanout[2] = 2
	V2IndexFanout[3] = 3

	for i := 3; i < len(V2IndexFanout); i++ {
		V2IndexFanout[i] = 3
	}

	fanout := make([]byte, FanoutWidth)
	for i, n := range V2IndexFanout {
		binary.BigEndian.PutUint32(fanout[i*FanoutEntryWidth:], n)
	}

	buf := make([]byte, 0, OffsetV2Start+3*(ObjectEntryV2Width)+ObjectLargeOffsetWidth)
	buf = append(buf, V2IndexHeader...)
	buf = append(buf, fanout...)
	buf = append(buf, V2IndexNames...)
	buf = append(buf, V2IndexCRCs...)
	buf = append(buf, V2IndexOffsets...)

	V2Index.f = bytes.NewReader(buf)
}
