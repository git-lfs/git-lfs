package pack

import (
	"compress/zlib"
	"io"
)

// ChainBase represents the "base" component of a delta-base chain.
type ChainBase struct {
	// offset returns the offset into the given io.ReaderAt where the read
	// will begin.
	offset int64
	// size is the total uncompressed size of the data in the base chain.
	size int64
	// typ is the type of data that this *ChainBase encodes.
	typ PackedObjectType

	// r is the io.ReaderAt yielding a stream of zlib-compressed data.
	r io.ReaderAt
}

// Unpack inflates and returns the uncompressed data encoded in the base
// element.
//
// If there was any error in reading the compressed data (invalid headers,
// etc.), it will be returned immediately.
func (b *ChainBase) Unpack() ([]byte, error) {
	zr, err := zlib.NewReader(&OffsetReaderAt{
		r: b.r,
		o: b.offset,
	})

	if err != nil {
		return nil, err
	}

	defer zr.Close()

	buf := make([]byte, b.size)
	if _, err := io.ReadFull(zr, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// ChainBase returns the type of the object it encodes.
func (b *ChainBase) Type() PackedObjectType {
	return b.typ
}
