package gitobj

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectTypeFromString(t *testing.T) {
	for str, typ := range map[string]ObjectType{
		"blob":           BlobObjectType,
		"tree":           TreeObjectType,
		"commit":         CommitObjectType,
		"tag":            TagObjectType,
		"something else": UnknownObjectType,
	} {
		t.Run(str, func(t *testing.T) {
			assert.Equal(t, typ, ObjectTypeFromString(str))
		})
	}
}

func TestObjectTypeToString(t *testing.T) {
	for typ, str := range map[ObjectType]string{
		BlobObjectType:            "blob",
		TreeObjectType:            "tree",
		CommitObjectType:          "commit",
		TagObjectType:             "tag",
		UnknownObjectType:         "unknown",
		ObjectType(math.MaxUint8): "<unknown>",
	} {
		t.Run(str, func(t *testing.T) {
			assert.Equal(t, str, typ.String())
		})
	}
}
