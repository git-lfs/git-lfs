package git

import (
	"io"

	"github.com/github/git-lfs/tools"
)

type PacketWriter struct {
	// buf is an internal buffer used to store data until enough has been
	// collected to write a full packet, or the buffer was instructed to
	// flush.
	buf []byte
	// proto is the place where packets get written.
	proto *protocol
}

var _ io.Writer = new(PacketWriter)

// NewPacketWriter returns a new *PacketWriter, which will write to the
// underlying data stream "w". The internal buffer is initialized with the given
// capacity, "c".
//
// If "w" is already a `*PacketWriter`, it will be returned as-is.
func NewPacketWriter(w io.Writer, c int) *PacketWriter {
	if pw, ok := w.(*PacketWriter); ok {
		return pw
	}

	return &PacketWriter{
		buf:   make([]byte, 0, c),
		proto: newProtocolRW(nil, w),
	}
}

// Write implements the io.Writer interface's `Write` method by providing a
// packet-based backend to the given buffer "p".
//
// As many bytes are removed from "p" as possible and stored in an internal
// buffer until the amount of data in the internal buffer is enough to write a
// single packet. Once the internal buffer is full, a packet is written to the
// underlying stream of data, and the process repeats.
//
// When the caller has no more data to write in the given chunk of packets, a
// subsequent call to `Write(p []byte)` MUST be made with a nil slice, to flush
// the remaining data in the buffer, and write the terminating bytes to the
// underlying packet stream.
//
// Write returns the number of bytes in "p" accepted into the writer, which
// _MAY_ be written to the underlying protocol stream, or may be written into
// the internal buffer.
//
// If any error was encountered while either buffering or writing, that
// error is returned, along with the number of bytes written to the underlying
// protocol stream, as described above.
func (w *PacketWriter) Write(p []byte) (int, error) {
	var n int

	if p == nil {
		// If we got an empty sequence of bytes, let's flush the data
		// stored in the buffer, and then write the a packet termination
		// sequence.

		if _, err := w.flush(); err != nil {
			return 0, err
		}

		if err := w.proto.writeFlush(); err != nil {
			return 0, err
		}
	}

	for len(p) > 0 {
		// While there is still data left to process in "p", grab as
		// much of it as we can while not allowing the internal buffer
		// to exceed the MaxPacketLength const.
		m := tools.MinInt(len(p), MaxPacketLength-len(w.buf))

		// Append on all of the data that we could into the internal
		// buffer.
		w.buf = append(w.buf, p[:m]...)
		// Truncate "p" such that it no longer includes the data that we
		// have in the internal buffer.
		p = p[m:]

		n = n + m

		if len(w.buf) == MaxPacketLength {
			// If we were able to grab an entire packet's worth of
			// data, flush the buffer.

			if _, err := w.flush(); err != nil {
				return n, err
			}

		}
	}

	return n, nil
}

// flush writes any data in the internal buffer out to the underlying protocol
// stream. If the amount of data in the internal buffer exceeds the
// MaxPacketLength, the data will be written in multiple packets to accommodate.
//
// flush returns the number of bytes written to the underlying packet stream,
// and any error that it encountered along the way.
func (w *PacketWriter) flush() (int, error) {
	var n int

	for len(w.buf) > 0 {
		if err := w.proto.writePacket(w.buf); err != nil {
			return 0, err
		}

		m := tools.MinInt(len(w.buf), MaxPacketLength)

		w.buf = w.buf[m:]

		n = n + m
	}

	return n, nil
}
