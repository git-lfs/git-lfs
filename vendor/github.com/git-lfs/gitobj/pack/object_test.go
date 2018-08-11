package pack

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectTypeReturnsObjectType(t *testing.T) {
	o := &Object{
		typ: TypeCommit,
	}

	assert.Equal(t, TypeCommit, o.Type())
}

func TestObjectUnpackUnpacksData(t *testing.T) {
	expected := []byte{0x1, 0x2, 0x3, 0x4}

	o := &Object{
		data: &ChainSimple{
			X: expected,
		},
	}

	data, err := o.Unpack()

	assert.Equal(t, expected, data)
	assert.NoError(t, err)
}

func TestObjectUnpackPropogatesErrors(t *testing.T) {
	expected := fmt.Errorf("gitobj/pack: testing")

	o := &Object{
		data: &ChainSimple{
			Err: expected,
		},
	}

	data, err := o.Unpack()

	assert.Nil(t, data)
	assert.Equal(t, expected, err)
}
