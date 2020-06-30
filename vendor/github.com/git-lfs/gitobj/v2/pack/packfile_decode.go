package pack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash"
	"io"
)

var (
	// packHeader is the expected header that begins all valid packfiles.
	packHeader = []byte{'P', 'A', 'C', 'K'}

	// errBadPackHeader is a sentinel error value returned when the given
	// pack header does not match the expected one.
	errBadPackHeader = errors.New("gitobj/pack: bad pack header")
)

// DecodePackfile opens the packfile given by the io.ReaderAt "r" for reading.
// It does not apply any delta-base chains, nor does it do reading otherwise
// beyond the header.
//
// If the header is malformed, or otherwise cannot be read, an error will be
// returned without a corresponding packfile.
func DecodePackfile(r io.ReaderAt, hash hash.Hash) (*Packfile, error) {
	header := make([]byte, 12)
	if _, err := r.ReadAt(header[:], 0); err != nil {
		return nil, err
	}

	if !bytes.HasPrefix(header, packHeader) {
		return nil, errBadPackHeader
	}

	version := binary.BigEndian.Uint32(header[4:])
	objects := binary.BigEndian.Uint32(header[8:])

	return &Packfile{
		Version: version,
		Objects: objects,

		r: r,
		hash: hash,
	}, nil
}
