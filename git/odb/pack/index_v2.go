package pack

import (
	"bytes"
	"encoding/binary"
)

const (
	// V2 is an instance of IndexVersion corresponding to the V2 index file
	// format.
	V2 IndexVersion = 2
)

// v2Search implements the IndexVersion.Search method for V2 packfiles.
func v2Search(idx *Index, name []byte, at int64) (*IndexEntry, int, error) {
	var sha [20]byte
	if _, err := idx.readAt(sha[:], v2ShaOffset(at)); err != nil {
		return nil, 0, err
	}

	cmp := bytes.Compare(name, sha[:])
	if cmp != 0 {
		return nil, cmp, nil
	}

	var offs [4]byte
	if _, err := idx.readAt(offs[:], v2SmallOffsetOffset(at, int64(idx.Count()))); err != nil {
		return nil, 0, err
	}

	loc := uint64(binary.BigEndian.Uint32(offs[:]))
	if loc&0x80000000 > 0 {
		// If the most significant bit (MSB) of the offset is set, then
		// the offset encodes the location for an 8-byte offset.
		//
		// Mask away (offs&0x7fffffff) the MSB to return the remaining
		// offset.
		var offs [8]byte
		if _, err := idx.readAt(offs[:], int64(loc&0x7fffffff)); err != nil {
			return nil, 0, err
		}

		loc = binary.BigEndian.Uint64(offs[:])
	}
	return &IndexEntry{PackOffset: loc}, 0, nil
}

// v2ShaOffset returns the offset of a SHA1 given at "at" in the V2 index file.
func v2ShaOffset(at int64) int64 {
	// Skip the packfile index header and the L1 fanout table.
	return indexOffsetV2Start +
		// Skip until the desired name in the sorted names table.
		(indexObjectNameWidth * at)
}

// v2SmallOffsetOffset returns the offset of an object's small (4-byte) offset
// given by "at".
func v2SmallOffsetOffset(at, total int64) int64 {
	// Skip the packfile index header and the L1 fanout table.
	return indexOffsetV2Start +
		// Skip the name table.
		(indexObjectNameWidth * total) +
		// Skip the CRC table.
		(indexObjectCRCWidth * total) +
		// Skip until the desired index in the small offsets table.
		(indexObjectSmallOffsetWidth * at)
}
