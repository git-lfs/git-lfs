package tq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMissingObjectErrorsAreRecognizable(t *testing.T) {
	err := newObjectMissingError("some-name", "some-oid").(*MalformedObjectError)

	assert.Equal(t, "some-name", err.Name)
	assert.Equal(t, "some-oid", err.Oid)
	assert.True(t, err.Missing())
}

func TestCorruptObjectErrorsAreRecognizable(t *testing.T) {
	err := newCorruptObjectError("some-name", "some-oid").(*MalformedObjectError)

	assert.Equal(t, "some-name", err.Name)
	assert.Equal(t, "some-oid", err.Oid)
	assert.True(t, err.Corrupt())
}
