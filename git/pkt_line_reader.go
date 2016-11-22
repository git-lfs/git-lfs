package git

import (
	"io"

	"github.com/git-lfs/git-lfs/tools"
)

type pktlineReader struct {
	pl *pktline

	buf []byte
}

var _ io.Reader = new(pktlineReader)

func (r *pktlineReader) Read(p []byte) (int, error) {
	var n int

	if len(r.buf) > 0 {
		// If there is data in the buffer, shift as much out of it and
		// into the given "p" as we can.
		n = tools.MinInt(len(p), len(r.buf))

		copy(p, r.buf[:n])
		r.buf = r.buf[n:]
	}

	// Loop and grab as many packets as we can in a given "run", until we
	// have either, a) overfilled the given buffer "p", or we have started
	// to internally buffer in "r.buf".
	for len(r.buf) == 0 {
		chunk, err := r.pl.readPacket()
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
		nn := tools.MinInt(len(chunk), len(p[n:]))

		// Move that amount into "p", from where we left off.
		copy(p[n:], chunk[:nn])
		// And move the rest into the buffer.
		r.buf = append(r.buf, chunk[nn:]...)

		// Mark that we have read "nn" bytes into "p"
		n += nn
	}

	return n, nil
}
