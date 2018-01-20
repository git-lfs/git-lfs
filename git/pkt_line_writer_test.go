package git

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPktlineWriterWritesPacketsShorterThanMaxPacketSize(t *testing.T) {
	var buf bytes.Buffer

	w := NewPktlineWriter(&buf, 0)
	assertWriterWrite(t, w, []byte("Hello, world!"), 13)
	assertWriterWrite(t, w, nil, 0)

	pl := newPktline(&buf, nil)
	assertPacketRead(t, pl, []byte("Hello, world!"))
	assertPacketRead(t, pl, nil)
}

func TestPktlineWriterWritesPacketsEqualToMaxPacketLength(t *testing.T) {
	big := make([]byte, MaxPacketLength)
	for i, _ := range big {
		big[i] = 1
	}

	// Make a copy so that we can drain the data inside of it
	p := make([]byte, MaxPacketLength)
	copy(p, big)

	var buf bytes.Buffer

	w := NewPktlineWriter(&buf, 0)
	assertWriterWrite(t, w, p, len(big))
	assertWriterWrite(t, w, nil, 0)

	pl := newPktline(&buf, nil)
	assertPacketRead(t, pl, big)
	assertPacketRead(t, pl, nil)
}

func TestPktlineWriterWritesMultiplePacketsLessThanMaxPacketLength(t *testing.T) {
	var buf bytes.Buffer

	w := NewPktlineWriter(&buf, 0)
	assertWriterWrite(t, w, []byte("first\n"), len("first\n"))
	assertWriterWrite(t, w, []byte("second"), len("second"))
	assertWriterWrite(t, w, nil, 0)

	pl := newPktline(&buf, nil)
	assertPacketRead(t, pl, []byte("first\nsecond"))
	assertPacketRead(t, pl, nil)
}

func TestPktlineWriterWritesMultiplePacketsGreaterThanMaxPacketLength(t *testing.T) {
	var buf bytes.Buffer

	b1 := make([]byte, MaxPacketLength*3/4)
	for i, _ := range b1 {
		b1[i] = 1
	}

	b2 := make([]byte, MaxPacketLength*3/4)
	for i, _ := range b2 {
		b2[i] = 2
	}

	w := NewPktlineWriter(&buf, 0)
	assertWriterWrite(t, w, b1, len(b1))
	assertWriterWrite(t, w, b2, len(b2))
	assertWriterWrite(t, w, nil, 0)

	// offs is how far into b2 we needed to buffer before writing an entire
	// packet
	offs := MaxPacketLength - len(b1)

	pl := newPktline(&buf, nil)
	assertPacketRead(t, pl, append(b1, b2[:offs]...))
	assertPacketRead(t, pl, b2[offs:])
	assertPacketRead(t, pl, nil)
}

func TestPktlineWriterAllowsFlushesOnNil(t *testing.T) {
	assert.NoError(t, (*PktlineWriter)(nil).Flush())
}

func TestPktlineWriterDoesntWrapItself(t *testing.T) {
	itself := &PktlineWriter{}
	nw := NewPktlineWriter(itself, 0)

	assert.Equal(t, itself, nw)
}

func assertWriterWrite(t *testing.T, w *PktlineWriter, p []byte, plen int) {
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
