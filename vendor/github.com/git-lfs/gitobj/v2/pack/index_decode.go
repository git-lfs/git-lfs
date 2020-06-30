package pack

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
)

const (
	// indexMagicWidth is the width of the magic header of packfiles version
	// 2 and newer.
	indexMagicWidth = 4
	// indexVersionWidth is the width of the version following the magic
	// header.
	indexVersionWidth = 4
	// indexV2Width is the total width of the header in V2.
	indexV2Width = indexMagicWidth + indexVersionWidth
	// indexV1Width is the total width of the header in V1.
	indexV1Width = 0

	// indexFanoutEntries is the number of entries in the fanout table.
	indexFanoutEntries = 256
	// indexFanoutEntryWidth is the width of each entry in the fanout table.
	indexFanoutEntryWidth = 4
	// indexFanoutWidth is the width of the entire fanout table.
	indexFanoutWidth = indexFanoutEntries * indexFanoutEntryWidth

	// indexOffsetV1Start is the location of the first object outside of the
	// V1 header.
	indexOffsetV1Start = indexV1Width + indexFanoutWidth
	// indexOffsetV2Start is the location of the first object outside of the
	// V2 header.
	indexOffsetV2Start = indexV2Width + indexFanoutWidth

	// indexObjectCRCWidth is the width of the CRC accompanying each object
	// in V2.
	indexObjectCRCWidth = 4
	// indexObjectSmallOffsetWidth is the width of the small offset encoded
	// into each object.
	indexObjectSmallOffsetWidth = 4
	// indexObjectLargeOffsetWidth is the width of the optional large offset
	// encoded into the small offset.
	indexObjectLargeOffsetWidth = 8
)

var (
	// ErrShortFanout is an error representing situations where the entire
	// fanout table could not be read, and is thus too short.
	ErrShortFanout = fmt.Errorf("gitobj/pack: too short fanout table")

	// indexHeader is the first four "magic" bytes of index files version 2
	// or newer.
	indexHeader = []byte{0xff, 0x74, 0x4f, 0x63}
)

// DecodeIndex decodes an index whose underlying data is supplied by "r".
//
// DecodeIndex reads only the header and fanout table, and does not eagerly
// parse index entries.
//
// If there was an error parsing, it will be returned immediately.
func DecodeIndex(r io.ReaderAt, hash hash.Hash) (*Index, error) {
	version, err := decodeIndexHeader(r, hash)
	if err != nil {
		return nil, err
	}

	fanout, err := decodeIndexFanout(r, version.Width())
	if err != nil {
		return nil, err
	}

	return &Index{
		version: version,
		fanout:  fanout,

		r: r,
	}, nil
}

// decodeIndexHeader determines which version the index given by "r" is.
func decodeIndexHeader(r io.ReaderAt, hash hash.Hash) (IndexVersion, error) {
	hdr := make([]byte, 4)
	if _, err := r.ReadAt(hdr, 0); err != nil {
		return nil, err
	}

	if bytes.Equal(hdr, indexHeader) {
		vb := make([]byte, 4)
		if _, err := r.ReadAt(vb, 4); err != nil {
			return nil, err
		}

		version := binary.BigEndian.Uint32(vb)
		switch version {
		case 1:
			return &V1{hash: hash}, nil
		case 2:
			return &V2{hash: hash}, nil
		}
		return nil, &UnsupportedVersionErr{uint32(version)}
	}
	return &V1{hash: hash}, nil
}

// decodeIndexFanout decodes the fanout table given by "r" and beginning at the
// given offset.
func decodeIndexFanout(r io.ReaderAt, offset int64) ([]uint32, error) {
	b := make([]byte, 256*4)
	if _, err := r.ReadAt(b, offset); err != nil {
		if err == io.EOF {
			return nil, ErrShortFanout
		}
		return nil, err
	}

	fanout := make([]uint32, 256)
	for i, _ := range fanout {
		fanout[i] = binary.BigEndian.Uint32(b[(i * 4):])
	}

	return fanout, nil
}
