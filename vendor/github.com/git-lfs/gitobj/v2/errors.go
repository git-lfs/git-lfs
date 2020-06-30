package gitobj

import "fmt"

// UnexpectedObjectType is an error type that represents a scenario where an
// object was requested of a given type "Wanted", and received as a different
// _other_ type, "Wanted".
type UnexpectedObjectType struct {
	// Got was the object type requested.
	Got ObjectType
	// Wanted was the object type received.
	Wanted ObjectType
}

// Error implements the error.Error() function.
func (e *UnexpectedObjectType) Error() string {
	return fmt.Sprintf("gitobj: unexpected object type, got: %q, wanted: %q", e.Got, e.Wanted)
}
