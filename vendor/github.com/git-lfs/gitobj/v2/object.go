package gitobj

import (
	"hash"
	"io"
)

// Object is an interface satisfied by any concrete type that represents a loose
// Git object.
type Object interface {
	// Encode takes an io.Writer, "to", and encodes an uncompressed
	// Git-compatible representation of itself to that stream.
	//
	// It must return "n", the number of uncompressed bytes written to that
	// stream, along with "err", any error that was encountered during the
	// write.
	//
	// Any error that was encountered should be treated as "fatal-local",
	// meaning that a particular invocation of Encode() cannot progress, and
	// an accurate number "n" of bytes written up that point should be
	// returned.
	Encode(to io.Writer) (n int, err error)

	// Decode takes an io.Reader, "from" as well as a size "size" (the
	// number of uncompressed bytes on the stream that represent the object
	// trying to be decoded) and decodes the encoded object onto itself,
	// as a mutative transaction.
	//
	// It returns the number of uncompressed bytes "n" that an invoication
	// of this function has advanced the io.Reader, "from", as well as any
	// error that was encountered along the way.
	//
	// If an(y) error was encountered, it should be returned immediately,
	// along with the number of bytes read up to that point.
	Decode(hash hash.Hash, from io.Reader, size int64) (n int, err error)

	// Type returns the ObjectType constant that represents an instance of
	// the implementing type.
	Type() ObjectType
}
