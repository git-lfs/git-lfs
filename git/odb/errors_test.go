package odb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnexpectedObjectTypeErrFormatting(t *testing.T) {
	err := &UnexpectedObjectType{
		Got: TreeObjectType, Wanted: BlobObjectType,
	}

	assert.Equal(t, "git/odb: unexpected object type, got: \"tree\", wanted: \"blob\"", err.Error())
}
