package pack

import (
	"fmt"

	"github.com/git-lfs/git-lfs/errors"
)

// IndexVersion is a constant type that represents the version of encoding used
// by a particular index version.
type IndexVersion uint32

const (
	// VersionUnknown is the zero-value for IndexVersion, and represents an
	// unknown version.
	VersionUnknown IndexVersion = 0
)

// Width returns the width of the header given in the respective version.
func (v IndexVersion) Width() int64 {
	switch v {
	case V2:
		return indexV2Width
	case V1:
		return indexV1Width
	}
	panic(fmt.Sprintf("git/odb/pack: width unknown for pack version %d", v))
}

var (
	// ErrIndexOutOfBounds is an error returned when the object lookup "at"
	// (see: Search() below) is out of bounds.
	ErrIndexOutOfBounds = errors.New("git/odb/pack: index is out of bounds")
)

// Search searches index "idx" for an object given by "name" at location "at".
//
// If will return the object if it was found, or a comparison determining
// whether to search above or below next.
//
// Otherwise, it will return an error.
func (v IndexVersion) Search(idx *Index, name []byte, at int64) (*IndexEntry, int, error) {
	if at > int64(idx.Count()) {
		return nil, 0, ErrIndexOutOfBounds
	}

	switch v {
	case V2:
		return v2Search(idx, name, at)
	case V1:
		return v1Search(idx, name, at)
	}
	return nil, 0, &UnsupportedVersionErr{Got: uint32(v)}
}
