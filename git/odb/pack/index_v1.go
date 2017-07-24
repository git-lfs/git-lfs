package pack

import (
	"bytes"
	"encoding/binary"
)

const (
	// V1 is an instance of IndexVersion corresponding to the V1 index file
	// format.
	V1 IndexVersion = 1
)

// v1Search implements the IndexVersion.Search method for V1 packfiles.
func v1Search(idx *Index, name []byte, at int64) (*IndexEntry, int, error) {
	var sha [20]byte
	if _, err := idx.readAt(sha[:], v1ShaOffset(at)); err != nil {
		return nil, 0, err
	}

	cmp := bytes.Compare(name, sha[:])
	if cmp != 0 {
		return nil, cmp, nil
	}

	var offs [4]byte
	if _, err := idx.readAt(offs[:], v1EntryOffset(at)); err != nil {
		return nil, 0, err
	}

	return &IndexEntry{
		PackOffset: uint64(binary.BigEndian.Uint32(offs[:])),
	}, 0, nil
}

// v1ShaOffset returns the location of the SHA1 of an object given at "at".
func v1ShaOffset(at int64) int64 {
	// Skip forward until the desired entry.
	return v1EntryOffset(at) +
		// Skip past the 4-byte object offset in the desired entry to
		// the SHA1.
		indexObjectSmallOffsetWidth
}

// v1EntryOffset returns the location of the packfile offset for the object
// given at "at".
func v1EntryOffset(at int64) int64 {
	// Skip the L1 fanout table
	return indexOffsetV1Start +
		// Skip the object entries before the one located at "at"
		(indexObjectEntryV1Width * at)
}
