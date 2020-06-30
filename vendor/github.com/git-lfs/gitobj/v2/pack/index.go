package pack

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
)

const maxHashSize = sha256.Size

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

	// r is the underlying set of encoded data comprising this index file.
	r io.ReaderAt
}

// Count returns the number of objects in the packfile.
func (i *Index) Count() int {
	return int(i.fanout[255])
}

// Close closes the packfile index if the underlying data stream is closeable.
// If so, it returns any error involved in closing.
func (i *Index) Close() error {
	if close, ok := i.r.(io.Closer); ok {
		return close.Close()
	}
	return nil
}

var (
	// errNotFound is an error returned by Index.Entry() (see: below) when
	// an object cannot be found in the index.
	errNotFound = fmt.Errorf("gitobj/pack: object not found in index")
)

// IsNotFound returns whether a given error represents a missing object in the
// index.
func IsNotFound(err error) bool {
	return err == errNotFound
}

// Entry returns an entry containing the offset of a given SHA1 "name".
//
// Entry operates in O(log(n))-time in the worst case, where "n" is the number
// of objects that begin with the first byte of "name".
//
// If the entry cannot be found, (nil, ErrNotFound) will be returned. If there
// was an error searching for or parsing an entry, it will be returned as (nil,
// err).
//
// Otherwise, (entry, nil) will be returned.
func (i *Index) Entry(name []byte) (*IndexEntry, error) {
	var last *bounds
	bounds := i.bounds(name)

	for bounds.Left() < bounds.Right() {
		if last.Equal(bounds) {
			// If the bounds are unchanged, that means either that
			// the object does not exist in the packfile, or the
			// fanout table is corrupt.
			//
			// Either way, we won't be able to find the object.
			// Return immediately to prevent infinite looping.
			return nil, errNotFound
		}
		last = bounds

		// Find the midpoint between the upper and lower bounds.
		mid := bounds.Left() + ((bounds.Right() - bounds.Left()) / 2)

		got, err := i.version.Name(i, mid)
		if err != nil {
			return nil, err
		}

		if cmp := bytes.Compare(name, got); cmp == 0 {
			// If "cmp" is zero, that means the object at that index
			// "at" had a SHA equal to the one given by name, and we
			// are done.
			return i.version.Entry(i, mid)
		} else if cmp < 0 {
			// If the comparison is less than 0, we searched past
			// the desired object, so limit the upper bound of the
			// search to the midpoint.
			bounds = bounds.WithRight(mid)
		} else if cmp > 0 {
			// Likewise, if the comparison is greater than 0, we
			// searched below the desired object. Modify the bounds
			// accordingly.
			bounds = bounds.WithLeft(mid)
		}

	}

	return nil, errNotFound
}

// readAt is a convenience method that allow reading into the underlying data
// source from other callers within this package.
func (i *Index) readAt(p []byte, at int64) (n int, err error) {
	return i.r.ReadAt(p, at)
}

// bounds returns the initial bounds for a given name using the fanout table to
// limit search results.
func (i *Index) bounds(name []byte) *bounds {
	var left, right int64

	if name[0] == 0 {
		// If the lower bound is 0, there are no objects before it,
		// start at the beginning of the index file.
		left = 0
	} else {
		// Otherwise, make the lower bound the slot before the given
		// object.
		left = int64(i.fanout[name[0]-1])
	}

	if name[0] == 255 {
		// As above, if the upper bound is the max byte value, make the
		// upper bound the last object in the list.
		right = int64(i.Count())
	} else {
		// Otherwise, make the upper bound the first object which is not
		// within the given slot.
		right = int64(i.fanout[name[0]+1])
	}

	return newBounds(left, right)
}
