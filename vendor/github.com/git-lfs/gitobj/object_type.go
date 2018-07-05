package gitobj

import "strings"

// ObjectType is a constant enumeration type for identifying the kind of object
// type an implementing instance of the Object interface is.
type ObjectType uint8

const (
	UnknownObjectType ObjectType = iota
	BlobObjectType
	TreeObjectType
	CommitObjectType
	TagObjectType
)

// ObjectTypeFromString converts from a given string to an ObjectType
// enumeration instance.
func ObjectTypeFromString(s string) ObjectType {
	switch strings.ToLower(s) {
	case "blob":
		return BlobObjectType
	case "tree":
		return TreeObjectType
	case "commit":
		return CommitObjectType
	case "tag":
		return TagObjectType
	default:
		return UnknownObjectType
	}
}

// String implements the fmt.Stringer interface and returns a string
// representation of the ObjectType enumeration instance.
func (t ObjectType) String() string {
	switch t {
	case UnknownObjectType:
		return "unknown"
	case BlobObjectType:
		return "blob"
	case TreeObjectType:
		return "tree"
	case CommitObjectType:
		return "commit"
	case TagObjectType:
		return "tag"
	}
	return "<unknown>"
}
