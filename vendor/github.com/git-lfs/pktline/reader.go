package pktline

import (
	"io"
)

type PktlineReader struct {
	pl *Pktline

	buf []byte

	eof bool
}

// NewPktlineReader returns a new *PktlineReader, which will read from the
// underlying data stream "r". The internal buffer is initialized with the given
// capacity, "c".
//
// If "r" is already a `*PktlineReader`, it will be returned as-is.
func NewPktlineReader(r io.Reader, c int) *PktlineReader {
	if pr, ok := r.(*PktlineReader); ok {
		return pr
	}

	return &PktlineReader{
		buf: make([]byte, 0, c),
		pl:  NewPktline(r, nil),
	}
}

// NewPktlineReaderFromPktline returns a new *PktlineReader based on the
// underlying *Pktline object.  The internal buffer is initialized with the
// given capacity, "c".
func NewPktlineReaderFromPktline(pl *Pktline, c int) *PktlineReader {
	return &PktlineReader{
		buf: make([]byte, 0, c),
		pl:  pl,
	}
}

func (r *PktlineReader) Read(p []byte) (int, error) {
	var n int

	if r.eof {
		return 0, io.EOF
	}

	if len(r.buf) > 0 {
		// If there is data in the buffer, shift as much out of it and
		// into the given "p" as we can.
		n = minInt(len(p), len(r.buf))

		copy(p, r.buf[:n])
		r.buf = r.buf[n:]
	}

	// Loop and grab as many packets as we can in a given "run", until we
	// have either, a) overfilled the given buffer "p", or we have started
	// to internally buffer in "r.buf".
	for len(r.buf) == 0 {
		chunk, err := r.pl.ReadPacket()
		if err != nil {
			return n, err
		}

		if len(chunk) == 0 {
			// If we got an empty chunk, then we know that we have
			// reached the end of processing for this particular
			// packet, so let's terminate.

			r.eof = true
			return n, io.EOF
		}

		// Figure out how much of the packet we can read into "p".
		nn := minInt(len(chunk), len(p[n:]))

		// Move that amount into "p", from where we left off.
		copy(p[n:], chunk[:nn])
		// And move the rest into the buffer.
		r.buf = append(r.buf, chunk[nn:]...)

		// Mark that we have read "nn" bytes into "p"
		n += nn
	}

	return n, nil
}

// Reset causes the reader to reset the end-of-file indicator and continue
// reading packets from the underlying reader.
func (r *PktlineReader) Reset() {
	r.eof = false
}
