package pack

import "fmt"

// UnsupportedVersionErr is a type implementing 'error' which indicates a
// the presence of an unsupported packfile version.
type UnsupportedVersionErr struct {
	// Got is the unsupported version that was detected.
	Got uint32
}

// Error implements 'error.Error()'.
func (u *UnsupportedVersionErr) Error() string {
	return fmt.Sprintf("gitobj/pack: unsupported version: %d", u.Got)
}
