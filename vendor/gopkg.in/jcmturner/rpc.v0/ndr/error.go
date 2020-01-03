package ndr

import "fmt"

// Malformed implements the error interface for malformed NDR encoding errors.
type Malformed struct {
	EText string
}

// Error implements the error interface on the Malformed struct.
func (e Malformed) Error() string {
	return fmt.Sprintf("malformed NDR steam: %s", e.EText)
}
