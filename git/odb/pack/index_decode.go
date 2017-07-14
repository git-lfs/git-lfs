package pack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
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

		return version, &UnsupportedVersionErr{uint32(version)}
	}
	return IndexVersion(0), nil
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
