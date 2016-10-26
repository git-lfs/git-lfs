package git

import (
	"bytes"
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
	rw := newProtocolRW(buf, nil)

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

func TestFilterProtocolReadsWholePackets(t *testing.T) {
	tc := &PacketReadTestCase{
		In: []byte{
			0x30, 0x30, 0x30, 0x38, // 0008 (hex. length)
			0x1, 0x2, 0x3, 0x4, // payload
		},
		Payload: []byte{0x1, 0x2, 0x3, 0x4},
	}

	tc.Assert(t)
}

func TestFilterProtocolDiscardsPacketsWithIncorrectLength(t *testing.T) {
	tc := &PacketReadTestCase{
		In: []byte{
			0x30, 0x30, 0x30, 0x34, // 0004 (hex. length)
			// No body
		},
		Err: "Invalid packet length.",
	}

	tc.Assert(t)
}

func TestFilterProtocolDiscardsPacketsWithUnparseableLength(t *testing.T) {
	tc := &PacketReadTestCase{
		In: []byte{
			0xff, 0xff, 0xff, 0xff, // ÿÿÿÿ (invalid hex. length)
			// No body
		},
		Err: "strconv.ParseInt: parsing \"\\xff\\xff\\xff\\xff\": invalid syntax",
	}

	tc.Assert(t)
}

func TestFilterProtocolDiscardsPacketsWithLengthZero(t *testing.T) {
	tc := &PacketReadTestCase{
		In: []byte{
			0x30, 0x30, 0x30, 0x30, // 0000 (hex. length)
			// Empty body
		},
	}

	tc.Assert(t)
}

func TestFilterProtocolReadsTextWithNewline(t *testing.T) {
	rw := newProtocolRW(bytes.NewReader([]byte{
		0x30, 0x30, 0x30, 0x39, // 0009 (hex. length)
		0x61, 0x62, 0x63, 0x64, 0xa,
		// Empty body
	}), nil)

	str, err := rw.readPacketText()

	assert.Nil(t, err)
	assert.Equal(t, "abcd", str)
}

func TestFilterProtocolReadsTextWithoutNewline(t *testing.T) {
	rw := newProtocolRW(bytes.NewReader([]byte{
		0x30, 0x30, 0x30, 0x38, // 0009 (hex. length)
		0x61, 0x62, 0x63, 0x64,
	}), nil)

	str, err := rw.readPacketText()

	assert.Nil(t, err)
	assert.Equal(t, "abcd", str)
}

func TestFilterProtocolReadsTextWithErr(t *testing.T) {
	rw := newProtocolRW(bytes.NewReader([]byte{
		0x30, 0x30, 0x30, 0x34, // 0004 (hex. length)
		// No body
	}), nil)

	str, err := rw.readPacketText()

	require.NotNil(t, err)
	assert.Equal(t, "Invalid packet length.", err.Error())
	assert.Equal(t, "", str)
}

func TestFilterProtocolAppendsPacketLists(t *testing.T) {
	rw := newProtocolRW(bytes.NewReader([]byte{
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

func TestFilterProtocolAppendsPacketListsAndReturnsErrs(t *testing.T) {
	rw := newProtocolRW(bytes.NewReader([]byte{
		0x30, 0x30, 0x30, 0x38, // 0009 (hex. length)
		0x61, 0x62, 0x63, 0x64, // "abcd"

		0x30, 0x30, 0x30, 0x34, // 0004 (hex. length)
		// No body
	}), nil)

	str, err := rw.readPacketList()

	require.NotNil(t, err)
	assert.Equal(t, "Invalid packet length.", err.Error())
	assert.Empty(t, str)
}

func TestFilterProtocolWritesPackets(t *testing.T) {
	var buf bytes.Buffer

	rw := newProtocolRW(nil, &buf)
	err := rw.writePacket([]byte{
		0x1, 0x2, 0x3, 0x4,
	})

	assert.Nil(t, err)
	assert.Equal(t, []byte{
		0x30, 0x30, 0x30, 0x38, // 0008 (hex. length)
		0x1, 0x2, 0x3, 0x4, // payload
	}, buf.Bytes())
}

func TestFilterProtocolDoesNotWritePacketsExceedingMaxLength(t *testing.T) {
	var buf bytes.Buffer

	rw := newProtocolRW(nil, &buf)
	err := rw.writePacket(make([]byte, MaxPacketLength+1))

	require.NotNil(t, err)
	assert.Equal(t, "Packet length exceeds maximal length", err.Error())
	assert.Empty(t, buf.Bytes())
}

func TestFilterProtocolWritesPacketText(t *testing.T) {
	var buf bytes.Buffer

	rw := newProtocolRW(nil, &buf)
	err := rw.writePacketText("abcd")

	assert.Nil(t, err)
	assert.Equal(t, []byte{
		0x30, 0x30, 0x30, 0x39, // 0009 (hex. length)
		0x61, 0x62, 0x63, 0x64, 0xa, // "abcd\n" (payload)
	}, buf.Bytes())
}

func TestFilterProtocolWritesPacketLists(t *testing.T) {
	var buf bytes.Buffer

	rw := newProtocolRW(nil, &buf)
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
