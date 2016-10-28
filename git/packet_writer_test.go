package git

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPacketWriterWritesPacketsShorterThanMaxPacketSize(t *testing.T) {
	var buf bytes.Buffer

	w := &PacketWriter{proto: newProtocolRW(nil, &buf)}
	assertWriterWrite(t, w, []byte("Hello, world!"), 0)
	assertWriterWrite(t, w, []byte{}, len("Hello, world!"))

	proto := newProtocolRW(&buf, nil)
	assertPacketRead(t, proto, []byte("Hello, world!"))
	assertPacketRead(t, proto, nil)
}

func TestPacketWriterWritesPacketsEqualToMaxPacketLength(t *testing.T) {
	big := make([]byte, MaxPacketLength)
	for i, _ := range big {
		big[i] = 1
	}

	// Make a copy so that we can drain the data inside of it
	p := make([]byte, MaxPacketLength)
	copy(p, big)

	var buf bytes.Buffer

	w := &PacketWriter{proto: newProtocolRW(nil, &buf)}
	assertWriterWrite(t, w, p, len(big))
	assertWriterWrite(t, w, []byte{}, 0)

	proto := newProtocolRW(&buf, nil)
	assertPacketRead(t, proto, big)
	assertPacketRead(t, proto, nil)
}

func TestPacketWriterWritesMultiplePacketsLessThanMaxPacketLength(t *testing.T) {
	var buf bytes.Buffer

	w := &PacketWriter{proto: newProtocolRW(nil, &buf)}
	assertWriterWrite(t, w, []byte("first\n"), 0)
	assertWriterWrite(t, w, []byte("second"), 0)
	assertWriterWrite(t, w, []byte{}, len("first\nsecond"))

	proto := newProtocolRW(&buf, nil)
	assertPacketRead(t, proto, []byte("first\nsecond"))
	assertPacketRead(t, proto, nil)
}

func TestPacketWriterWritesMultiplePacketsGreaterThanMaxPacketLength(t *testing.T) {
	var buf bytes.Buffer

	b1 := make([]byte, MaxPacketLength*3/4)
	p1 := make([]byte, len(b1))
	for i, _ := range b1 {
		b1[i] = 1
	}
	copy(p1, b1)

	b2 := make([]byte, MaxPacketLength*3/4)
	p2 := make([]byte, len(b2))
	for i, _ := range b2 {
		b2[i] = 1
	}
	copy(p2, b1)

	w := &PacketWriter{proto: newProtocolRW(nil, &buf)}
	assertWriterWrite(t, w, p1, 0)
	assertWriterWrite(t, w, p2, MaxPacketLength)
	assertWriterWrite(t, w, []byte{}, (len(b1)+len(b2))-MaxPacketLength)

	// offs is how far into b2 we needed to buffer before writing an entire
	// packet
	offs := MaxPacketLength - len(b1)

	proto := newProtocolRW(&buf, nil)
	assertPacketRead(t, proto, append(b1, b2[:offs]...))
	assertPacketRead(t, proto, b2[offs:])
	assertPacketRead(t, proto, nil)
}

func assertWriterWrite(t *testing.T, w io.Writer, p []byte, plen int) {
	n, err := w.Write(p)

	assert.Nil(t, err)
	assert.Equal(t, plen, n)
}

func assertPacketRead(t *testing.T, proto *protocol, expected []byte) {
	got, err := proto.readPacket()

	assert.Nil(t, err)
	assert.Equal(t, expected, got)
}
