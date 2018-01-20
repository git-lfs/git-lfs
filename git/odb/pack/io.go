package pack

import "io"

// OffsetReaderAt transforms an io.ReaderAt into an io.Reader by beginning and
// advancing all reads at the given offset.
type OffsetReaderAt struct {
	// r is the data source for this instance of *OffsetReaderAt.
	r io.ReaderAt

	// o if the number of bytes read from the underlying data source, "r".
	// It is incremented upon reads.
	o int64
}

// Read implements io.Reader.Read by reading into the given []byte, "p" from the
// last known offset provided to the OffsetReaderAt.
//
// It returns any error encountered from the underlying data stream, and
// advances the reader forward by "n", the number of bytes read from the
// underlying data stream.
func (r *OffsetReaderAt) Read(p []byte) (n int, err error) {
	n, err = r.r.ReadAt(p, r.o)
	r.o += int64(n)

	return n, err
}
