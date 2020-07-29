package pack

import (
	"errors"
	"fmt"
)

// PackedObjectType is a constant type that is defined for all valid object
// types that a packed object can represent.
type PackedObjectType uint8

const (
	// TypeNone is the zero-value for PackedObjectType, and represents the
	// absence of a type.
	TypeNone PackedObjectType = iota
	// TypeCommit is the PackedObjectType for commit objects.
	TypeCommit
	// TypeTree is the PackedObjectType for tree objects.
	TypeTree
	// Typeblob is the PackedObjectType for blob objects.
	TypeBlob
	// TypeTag is the PackedObjectType for tag objects.
	TypeTag

	// TypeObjectOffsetDelta is the type for OBJ_OFS_DELTA-typed objects.
	TypeObjectOffsetDelta PackedObjectType = 6
	// TypeObjectReferenceDelta is the type for OBJ_REF_DELTA-typed objects.
	TypeObjectReferenceDelta PackedObjectType = 7
)

// String implements fmt.Stringer and returns an encoding of the type valid for
// use in the loose object format protocol (see: package 'gitobj' for more).
//
// If the receiving instance is not defined, String() will panic().
func (t PackedObjectType) String() string {
	switch t {
	case TypeNone:
		return "<none>"
	case TypeCommit:
		return "commit"
	case TypeTree:
		return "tree"
	case TypeBlob:
		return "blob"
	case TypeTag:
		return "tag"
	case TypeObjectOffsetDelta:
		return "obj_ofs_delta"
	case TypeObjectReferenceDelta:
		return "obj_ref_delta"
	}

	panic(fmt.Sprintf("gitobj/pack: unknown object type: %d", t))
}

var (
	errUnrecognizedObjectType = errors.New("gitobj/pack: unrecognized object type")
)
