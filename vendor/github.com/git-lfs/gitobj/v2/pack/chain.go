package pack

// Chain represents an element in the delta-base chain corresponding to a packed
// object.
type Chain interface {
	// Unpack unpacks the data encoded in the delta-base chain up to and
	// including the receiving Chain implementation by applying the
	// delta-base chain successively to itself.
	//
	// If there was an error in the delta-base resolution, i.e., the chain
	// is malformed, has a bad instruction, or there was a file read error, this
	// function is expected to return that error.
	//
	// In the event that a non-nil error is returned, it is assumed that the
	// unpacked data this function returns is malformed, or otherwise
	// corrupt.
	Unpack() ([]byte, error)

	// Type returns the type of the receiving chain element.
	Type() PackedObjectType
}
