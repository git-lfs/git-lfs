package git

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type PacketReadTestCase struct {
	In []byte

	Payload []byte
	Err     string
}

func (c *PacketReadTestCase) Assert(t *testing.T) {
	buf := bytes.NewReader(c.In)
	rw := newPktline(buf, nil)

	pkt, err := rw.readPacket()

	if len(c.Payload) > 0 {
		assert.Equal(t, c.Payload, pkt)
	} else {
		assert.Empty(t, pkt)
	}

	if len(c.Err) > 0 {
		require.NotNil(t, err)
		assert.Equal(t, c.Err, err.Error())
	} else {
		assert.Nil(t, err)
	}
}

func TestPktLineReadsWholePackets(t *testing.T) {
	tc := &PacketReadTestCase{
		In: []byte{
			0x30, 0x30, 0x30, 0x38, // 0008 (hex. length)
			0x1, 0x2, 0x3, 0x4, // payload
		},
		Payload: []byte{0x1, 0x2, 0x3, 0x4},
	}

	tc.Assert(t)
}

func TestPktLineNoPacket(t *testing.T) {
	tc := &PacketReadTestCase{
		In:  []byte{},
		Err: io.EOF.Error(),
	}

	tc.Assert(t)
}

func TestPktLineEmptyPacket(t *testing.T) {
	tc := &PacketReadTestCase{
		In: []byte{
			0x30, 0x30, 0x30, 0x34,
			// No body (invalid)
		},

		Err: "invalid packet length",
	}

	tc.Assert(t)

}

func TestPktLineFlushPacket(t *testing.T) {
	tc := &PacketReadTestCase{
		In: []byte{0x30, 0x30, 0x30, 0x30}, // Flush packet

		Payload: []byte{},
		Err:     "",
	}

	tc.Assert(t)
}

func TestPktLineDiscardsPacketsWithUnparseableLength(t *testing.T) {
	tc := &PacketReadTestCase{
		In: []byte{
			0xff, 0xff, 0xff, 0xff, // ÿÿÿÿ (invalid hex. length)
			// No body
		},
		Err: "strconv.ParseInt: parsing \"\\xff\\xff\\xff\\xff\": invalid syntax",
	}

	tc.Assert(t)
}

func TestPktLineReadsTextWithNewline(t *testing.T) {
	rw := newPktline(bytes.NewReader([]byte{
		0x30, 0x30, 0x30, 0x39, // 0009 (hex. length)
		0x61, 0x62, 0x63, 0x64, 0xa,
		// Empty body
	}), nil)

	str, err := rw.readPacketText()

	assert.Nil(t, err)
	assert.Equal(t, "abcd", str)
}

func TestPktLineReadsTextWithoutNewline(t *testing.T) {
	rw := newPktline(bytes.NewReader([]byte{
		0x30, 0x30, 0x30, 0x38, // 0009 (hex. length)
		0x61, 0x62, 0x63, 0x64,
	}), nil)

	str, err := rw.readPacketText()

	assert.Nil(t, err)
	assert.Equal(t, "abcd", str)
}

func TestPktLineReadsTextWithErr(t *testing.T) {
	rw := newPktline(bytes.NewReader([]byte{
		0x30, 0x30, 0x30, 0x34, // 0004 (hex. length)
		// No body
	}), nil)

	str, err := rw.readPacketText()

	require.NotNil(t, err)
	assert.Equal(t, "invalid packet length", err.Error())
	assert.Equal(t, "", str)
}

func TestPktLineAppendsPacketLists(t *testing.T) {
	rw := newPktline(bytes.NewReader([]byte{
		0x30, 0x30, 0x30, 0x38, // 0009 (hex. length)
		0x61, 0x62, 0x63, 0x64, // "abcd"

		0x30, 0x30, 0x30, 0x38, // 0008 (hex. length)
		0x65, 0x66, 0x67, 0x68, // "efgh"

		0x30, 0x30, 0x30, 0x30, // 0000 (hex. length)
	}), nil)

	str, err := rw.readPacketList()

	assert.Nil(t, err)
	assert.Equal(t, []string{"abcd", "efgh"}, str)
}

func TestPktLineAppendsPacketListsAndReturnsErrs(t *testing.T) {
	rw := newPktline(bytes.NewReader([]byte{
		0x30, 0x30, 0x30, 0x38, // 0009 (hex. length)
		0x61, 0x62, 0x63, 0x64, // "abcd"

		0x30, 0x30, 0x30, 0x34, // 0004 (hex. length)
		// No body
	}), nil)

	str, err := rw.readPacketList()

	require.NotNil(t, err)
	assert.Equal(t, "invalid packet length", err.Error())
	assert.Empty(t, str)
}

func TestPktLineWritesPackets(t *testing.T) {
	var buf bytes.Buffer

	rw := newPktline(nil, &buf)
	require.Nil(t, rw.writePacket([]byte{
		0x1, 0x2, 0x3, 0x4,
	}))
	require.Nil(t, rw.writeFlush())

	assert.Equal(t, []byte{
		0x30, 0x30, 0x30, 0x38, // 0008 (hex. length)
		0x1, 0x2, 0x3, 0x4, // payload
		0x30, 0x30, 0x30, 0x30, // 0000 (flush packet)
	}, buf.Bytes())
}

func TestPktLineWritesPacketsEqualToMaxLength(t *testing.T) {
	var buf bytes.Buffer

	rw := newPktline(nil, &buf)
	err := rw.writePacket(make([]byte, MaxPacketLength))

	assert.Nil(t, err)
	assert.Equal(t, 4+MaxPacketLength, len(buf.Bytes()))
}

func TestPktLineDoesNotWritePacketsExceedingMaxLength(t *testing.T) {
	var buf bytes.Buffer

	rw := newPktline(nil, &buf)
	err := rw.writePacket(make([]byte, MaxPacketLength+1))

	require.NotNil(t, err)
	assert.Equal(t, "packet length exceeds maximal length", err.Error())
	assert.Empty(t, buf.Bytes())
}

func TestPktLineWritesPacketText(t *testing.T) {
	var buf bytes.Buffer

	rw := newPktline(nil, &buf)

	require.Nil(t, rw.writePacketText("abcd"))
	require.Nil(t, rw.writeFlush())

	assert.Equal(t, []byte{
		0x30, 0x30, 0x30, 0x39, // 0009 (hex. length)
		0x61, 0x62, 0x63, 0x64, 0xa, // "abcd\n" (payload)
		0x30, 0x30, 0x30, 0x30, // 0000 (flush packet)
	}, buf.Bytes())
}

func TestPktLineWritesPacketLists(t *testing.T) {
	var buf bytes.Buffer

	rw := newPktline(nil, &buf)
	err := rw.writePacketList([]string{"foo", "bar"})

	assert.Nil(t, err)
	assert.Equal(t, []byte{
		0x30, 0x30, 0x30, 0x38, // 0008 (hex. length)
		0x66, 0x6f, 0x6f, 0xa, // "foo\n" (payload)

		0x30, 0x30, 0x30, 0x38, // 0008 (hex. length)
		0x62, 0x61, 0x72, 0xa, // "bar\n" (payload)

		0x30, 0x30, 0x30, 0x30, // 0000 (hex. length)
	}, buf.Bytes())
}
