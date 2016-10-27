package git

import (
	"io"
	"math"
)

type packetReader struct {
	proto *protocol

	buf []byte
}

var _ io.Reader = new(packetReader)

func (r *packetReader) Read(p []byte) (int, error) {
	var n int

	if len(r.buf) > 0 {
		// If there is data in the buffer, shift as much out of it and
		// into the given "p" as we can.
		n = int(math.Min(float64(len(p)), float64(len(r.buf))))

		copy(p, r.buf[:n])
		r.buf = r.buf[n:]
	}

	// Loop and grab as many packets as we can in a given "run", until we
	// have either, a) overfilled the given buffer "p", or we have started
	// to internally buffer in "r.buf".
	for len(r.buf) == 0 {
		chunk, err := r.proto.readPacket()
		if err != nil {
			return n, err
		}

		if len(chunk) == 0 {
			// If we got an empty chunk, then we know that we have
			// reached the end of processing for this particular
			// packet, so let's terminate.

			return n, io.EOF
		}

		// Figure out how much of the packet we can read into "p".
		nn := int(math.Min(float64(len(chunk)), float64(len(p))))

		// Move that amount into "p", from where we left off.
		copy(p[n:], chunk[:nn])
		// And move the rest into the buffer.
		r.buf = append(r.buf, chunk[nn:]...)

		// Mark that we have read "nn" bytes into "p"
		n += nn

		if n >= len(p) {
			break
		}
	}

	return n, nil
}
