package pack

import (
	"encoding/binary"
	"hash"
)

// V2 implements IndexVersion for v2 packfiles.
type V2 struct {
	hash hash.Hash
}

// Name implements IndexVersion.Name by returning the 20 byte SHA-1 object name
// for the given entry at offset "at" in the v2 index file "idx".
func (v *V2) Name(idx *Index, at int64) ([]byte, error) {
	var sha [maxHashSize]byte

	hashlen := v.hash.Size()

	if _, err := idx.readAt(sha[:hashlen], v2ShaOffset(at, int64(hashlen))); err != nil {
		return nil, err
	}

	return sha[:hashlen], nil
}

// Entry implements IndexVersion.Entry for v2 packfiles by parsing and returning
// the IndexEntry specified at the offset "at" in the given index file.
func (v *V2) Entry(idx *Index, at int64) (*IndexEntry, error) {
	var offs [4]byte

	hashlen := v.hash.Size()

	if _, err := idx.readAt(offs[:], v2SmallOffsetOffset(at, int64(idx.Count()), int64(hashlen))); err != nil {
		return nil, err
	}

	loc := uint64(binary.BigEndian.Uint32(offs[:]))
	if loc&0x80000000 > 0 {
		// If the most significant bit (MSB) of the offset is set, then
		// the offset encodes the indexed location for an 8-byte offset.
		//
		// Mask away (offs&0x7fffffff) the MSB to use as an index to
		// find the offset of the 8-byte pack offset.
		lo := v2LargeOffsetOffset(int64(loc&0x7fffffff), int64(idx.Count()), int64(hashlen))

		var offs [8]byte
		if _, err := idx.readAt(offs[:], lo); err != nil {
			return nil, err
		}

		loc = binary.BigEndian.Uint64(offs[:])
	}
	return &IndexEntry{PackOffset: loc}, nil
}

// Width implements IndexVersion.Width() by returning the number of bytes that
// v2 packfile index header occupy.
func (v *V2) Width() int64 {
	return indexV2Width
}

// v2ShaOffset returns the offset of a SHA1 given at "at" in the V2 index file.
func v2ShaOffset(at int64, hashlen int64) int64 {
	// Skip the packfile index header and the L1 fanout table.
	return indexOffsetV2Start +
		// Skip until the desired name in the sorted names table.
		(hashlen * at)
}

// v2SmallOffsetOffset returns the offset of an object's small (4-byte) offset
// given by "at".
func v2SmallOffsetOffset(at, total, hashlen int64) int64 {
	// Skip the packfile index header and the L1 fanout table.
	return indexOffsetV2Start +
		// Skip the name table.
		(hashlen * total) +
		// Skip the CRC table.
		(indexObjectCRCWidth * total) +
		// Skip until the desired index in the small offsets table.
		(indexObjectSmallOffsetWidth * at)
}

// v2LargeOffsetOffset returns the offset of an object's large (4-byte) offset,
// given by the index "at".
func v2LargeOffsetOffset(at, total, hashlen int64) int64 {
	// Skip the packfile index header and the L1 fanout table.
	return indexOffsetV2Start +
		// Skip the name table.
		(hashlen * total) +
		// Skip the CRC table.
		(indexObjectCRCWidth * total) +
		// Skip the small offsets table.
		(indexObjectSmallOffsetWidth * total) +
		// Seek to the large offset within the large offset(s) table.
		(indexObjectLargeOffsetWidth * at)
}
