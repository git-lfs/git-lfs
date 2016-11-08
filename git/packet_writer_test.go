package git

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPacketWriterWritesPacketsShorterThanMaxPacketSize(t *testing.T) {
	var buf bytes.Buffer

	w := NewPacketWriter(&buf, 0)
	assertWriterWrite(t, w, []byte("Hello, world!"), 13)
	assertWriterWrite(t, w, nil, 0)

	pl := newPktline(&buf, nil)
	assertPacketRead(t, pl, []byte("Hello, world!"))
	assertPacketRead(t, pl, nil)
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

	w := NewPacketWriter(&buf, 0)
	assertWriterWrite(t, w, p, len(big))
	assertWriterWrite(t, w, nil, 0)

	pl := newPktline(&buf, nil)
	assertPacketRead(t, pl, big)
	assertPacketRead(t, pl, nil)
}

func TestPacketWriterWritesMultiplePacketsLessThanMaxPacketLength(t *testing.T) {
	var buf bytes.Buffer

	w := NewPacketWriter(&buf, 0)
	assertWriterWrite(t, w, []byte("first\n"), len("first\n"))
	assertWriterWrite(t, w, []byte("second"), len("second"))
	assertWriterWrite(t, w, nil, 0)

	pl := newPktline(&buf, nil)
	assertPacketRead(t, pl, []byte("first\nsecond"))
	assertPacketRead(t, pl, nil)
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

	w := NewPacketWriter(&buf, 0)
	assertWriterWrite(t, w, p1, len(p1))
	assertWriterWrite(t, w, p2, len(p2))
	assertWriterWrite(t, w, nil, 0)

	// offs is how far into b2 we needed to buffer before writing an entire
	// packet
	offs := MaxPacketLength - len(b1)

	pl := newPktline(&buf, nil)
	assertPacketRead(t, pl, append(b1, b2[:offs]...))
	assertPacketRead(t, pl, b2[offs:])
	assertPacketRead(t, pl, nil)
}

func TestPacketWriterDoesntWrapItself(t *testing.T) {
	itself := &PacketWriter{}
	nw := NewPacketWriter(itself, 0)

	assert.Equal(t, itself, nw)
}

func assertWriterWrite(t *testing.T, w *PacketWriter, p []byte, plen int) {
	var n int
	var err error

	if p == nil {
		err = w.Flush()
	} else {
		n, err = w.Write(p)
	}

	assert.Nil(t, err)
	assert.Equal(t, plen, n)
}

func assertPacketRead(t *testing.T, pl *pktline, expected []byte) {
	got, err := pl.readPacket()

	assert.Nil(t, err)
	assert.Equal(t, expected, got)
}
