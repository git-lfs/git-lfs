package pack

import (
	"fmt"
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
	panic(fmt.Sprintf("git/odb/pack: width unknown for pack version %d", v))
}
