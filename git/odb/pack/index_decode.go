package pack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const (
	// V1Width is the total width of the header in V1.
	V1Width = 0

	// FanoutEntries is the number of entries in the fanout table.
	FanoutEntries = 256
	// FanoutEntryWidth is the width of each entry in the fanout table.
	FanoutEntryWidth = 4
	// FanoutWidth is the width of the entire fanout table.
	FanoutWidth = FanoutEntries * FanoutEntryWidth

	// OffsetV1Start is the location of the first object outside of the V1
	// header.
	OffsetV1Start = V1Width + FanoutWidth

	// ObjectNameWidth is the width of a SHA1 object name.
	ObjectNameWidth = 20
	// ObjectSmallOffsetWidth is the width of the small offset encoded into
	// each object.
	ObjectSmallOffsetWidth = 4

	// ObjectEntryV1Width is the width of one contiguous object entry in V1.
	ObjectEntryV1Width = ObjectNameWidth + ObjectSmallOffsetWidth
)

var (
	// ErrShortFanout is an error representing situations where the entire
	// fanout table could not be read, and is thus too short.
	ErrShortFanout = errors.New("git/odb/pack: too short fanout table")

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
func DecodeIndex(r io.ReaderAt) (*Index, error) {
	version, err := decodeIndexHeader(r)
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

		f: r,
	}, nil
}

// decodeIndexHeader determines which version the index given by "r" is.
func decodeIndexHeader(r io.ReaderAt) (IndexVersion, error) {
	hdr := make([]byte, 4)
	if _, err := r.ReadAt(hdr, 0); err != nil {
		return VersionUnknown, err
	}

	if bytes.Equal(hdr, indexHeader) {
		vb := make([]byte, 4)
		if _, err := r.ReadAt(vb, 4); err != nil {
			return VersionUnknown, err
		}

		version := IndexVersion(binary.BigEndian.Uint32(vb))
		switch version {
		case V1:
			return version, nil
		}

		return version, &UnsupportedVersionErr{uint32(version)}
	}
	return V1, nil
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
