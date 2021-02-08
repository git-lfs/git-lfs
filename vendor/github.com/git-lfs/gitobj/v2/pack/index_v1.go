package pack

import (
	"encoding/binary"
	"hash"
)

// V1 implements IndexVersion for v1 packfiles.
type V1 struct {
	hash hash.Hash
}

// Name implements IndexVersion.Name by returning the 20 byte SHA-1 object name
// for the given entry at offset "at" in the v1 index file "idx".
func (v *V1) Name(idx *Index, at int64) ([]byte, error) {
	var sha [MaxHashSize]byte

	hashlen := v.hash.Size()

	if _, err := idx.readAt(sha[:hashlen], v1ShaOffset(at, int64(hashlen))); err != nil {
		return nil, err
	}

	return sha[:hashlen], nil
}

// Entry implements IndexVersion.Entry for v1 packfiles by parsing and returning
// the IndexEntry specified at the offset "at" in the given index file.
func (v *V1) Entry(idx *Index, at int64) (*IndexEntry, error) {
	var offs [4]byte
	if _, err := idx.readAt(offs[:], v1EntryOffset(at, int64(v.hash.Size()))); err != nil {
		return nil, err
	}

	return &IndexEntry{
		PackOffset: uint64(binary.BigEndian.Uint32(offs[:])),
	}, nil
}

// Width implements IndexVersion.Width() by returning the number of bytes that
// v1 packfile index header occupy.
func (v *V1) Width() int64 {
	return indexV1Width
}

// v1ShaOffset returns the location of the SHA1 of an object given at "at".
func v1ShaOffset(at int64, hashlen int64) int64 {
	// Skip forward until the desired entry.
	return v1EntryOffset(at, hashlen) +
		// Skip past the 4-byte object offset in the desired entry to
		// the SHA1.
		indexObjectSmallOffsetWidth
}

// v1EntryOffset returns the location of the packfile offset for the object
// given at "at".
func v1EntryOffset(at int64, hashlen int64) int64 {
	// Skip the L1 fanout table
	return indexOffsetV1Start +
		// Skip the object entries before the one located at "at"
		((hashlen + indexObjectSmallOffsetWidth) * at)
}
