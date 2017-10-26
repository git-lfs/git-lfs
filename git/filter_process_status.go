package git

import "fmt"

// FilterProcessStatus is a constant type representing the various valid
// responses for `status=` in the Git filtering process protocol.
type FilterProcessStatus uint8

const (
	// StatusSuccess is a valid response when a successful event has
	// occurred.
	StatusSuccess FilterProcessStatus = iota + 1
	// StatusDelay is a valid response when a delay has occurred.
	StatusDelay
	// StatusError is a valid response when an error has occurred.
	StatusError
)

// String implements fmt.Stringer by returning a protocol-compliant
// representation of the receiving status, or panic()-ing if the Status is
// unknown.
func (s FilterProcessStatus) String() string {
	switch s {
	case StatusSuccess:
		return "success"
	case StatusDelay:
		return "delayed"
	case StatusError:
		return "error"
	}

	panic(fmt.Sprintf("git: unknown FilterProcessStatus '%d'", s))
}
