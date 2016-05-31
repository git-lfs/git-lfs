package gojsonschema

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUUIDFormatCheckerIsFormat(t *testing.T) {
	checker := UUIDFormatChecker{}

	assert.True(t, checker.IsFormat("01234567-89ab-cdef-0123-456789abcdef"))
	assert.True(t, checker.IsFormat("f1234567-89ab-cdef-0123-456789abcdef"))

	assert.False(t, checker.IsFormat("not-a-uuid"))
	assert.False(t, checker.IsFormat("g1234567-89ab-cdef-0123-456789abcdef"))
}
