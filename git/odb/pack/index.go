package pack

import (
	"io"
)

// Index stores information about the location of objects in a corresponding
// packfile.
type Index struct {
	// version is the encoding version used by this index.
	//
	// Currently, versions 1 and 2 are supported.
	version IndexVersion
	// fanout is the L1 fanout table stored in this index. For a given index
	// "i" into the array, the value stored at that index specifies the
	// number of objects in the packfile/index that are lexicographically
	// less than or equal to that index.
	//
	// See: https://github.com/git/git/blob/v2.13.0/Documentation/technical/pack-format.txt#L41-L45
	fanout []uint32

	// f is the underlying set of encoded data comprising this index file.
	f io.ReaderAt
}

// Count returns the number of objects in the packfile.
func (i *Index) Count() int {
	return int(i.fanout[255])
}

// readAt is a convenience method that allow reading into the underlying data
// source from other callers within this package.
func (i *Index) readAt(p []byte, at int64) (n int, err error) {
	return i.f.ReadAt(p, at)
}
