package git

import (
	"io"

	"github.com/git-lfs/git-lfs/tools"
)

// PktlineWriter is an implementation of `io.Writer` which writes data buffers
// "p" to an underlying pkt-line stream for use with the Git pkt-line format.
type PktlineWriter struct {
	// buf is an internal buffer used to store data until enough has been
	// collected to write a full packet, or the buffer was instructed to
	// flush.
	buf []byte
	// pl is the place where packets get written.
	pl *pktline
}

var _ io.Writer = new(PktlineWriter)

// NewPktlineWriter returns a new *PktlineWriter, which will write to the
// underlying data stream "w". The internal buffer is initialized with the given
// capacity, "c".
//
// If "w" is already a `*PktlineWriter`, it will be returned as-is.
func NewPktlineWriter(w io.Writer, c int) *PktlineWriter {
	if pw, ok := w.(*PktlineWriter); ok {
		return pw
	}

	return &PktlineWriter{
		buf: make([]byte, 0, c),
		pl:  newPktline(nil, w),
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
// subsequent call to `Flush()` SHOULD be made in order to signify that the
// current pkt sequence has terminated, and a new one can begin.
//
// Write returns the number of bytes in "p" accepted into the writer, which
// _MAY_ be written to the underlying protocol stream, or may be written into
// the internal buffer.
//
// If any error was encountered while either buffering or writing, that
// error is returned, along with the number of bytes written to the underlying
// protocol stream, as described above.
func (w *PktlineWriter) Write(p []byte) (int, error) {
	var n int

	for len(p[n:]) > 0 {
		// While there is still data left to process in "p", grab as
		// much of it as we can while not allowing the internal buffer
		// to exceed the MaxPacketLength const.
		m := tools.MinInt(len(p[n:]), MaxPacketLength-len(w.buf))

		// Append on all of the data that we could into the internal
		// buffer.
		w.buf = append(w.buf, p[n:n+m]...)

		n += m

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

// Flush empties the internal buffer used to store data temporarily and then
// writes the pkt-line's FLUSH packet, to signal that it is done writing this
// chunk of data.
func (w *PktlineWriter) Flush() error {
	if w == nil {
		return nil
	}

	if _, err := w.flush(); err != nil {
		return err
	}

	if err := w.pl.writeFlush(); err != nil {
		return err
	}

	return nil
}

// flush writes any data in the internal buffer out to the underlying protocol
// stream. If the amount of data in the internal buffer exceeds the
// MaxPacketLength, the data will be written in multiple packets to accommodate.
//
// flush returns the number of bytes written to the underlying packet stream,
// and any error that it encountered along the way.
func (w *PktlineWriter) flush() (int, error) {
	var n int

	for len(w.buf) > 0 {
		if err := w.pl.writePacket(w.buf); err != nil {
			return 0, err
		}

		m := tools.MinInt(len(w.buf), MaxPacketLength)

		w.buf = w.buf[m:]

		n = n + m
	}

	return n, nil
}
