package pack

// Object is an encapsulation of an object found in a packfile, or a packed
// object.
type Object struct {
	// data is the front-most element of the delta-base chain, and when
	// resolved, yields the uncompressed data of this object.
	data Chain
	// typ is the underlying object's type. It is not the type of the
	// front-most chain element, rather, the type of the actual object.
	typ PackedObjectType
}

// Unpack resolves the delta-base chain and returns an uncompressed, unpacked,
// and full representation of the data encoded by this object.
//
// If there was any error in unpacking this object, it is returned immediately,
// and the object's data can be assumed to be corrupt.
func (o *Object) Unpack() ([]byte, error) {
	return o.data.Unpack()
}

// Type returns the underlying object's type. Rather than the type of the
// front-most delta-base component, it is the type of the object itself.
func (o *Object) Type() PackedObjectType {
	return o.typ
}
