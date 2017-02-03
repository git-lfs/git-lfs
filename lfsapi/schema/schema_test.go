package schema

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ValidSchemaPath   = "fixture/valid.json"
	InvalidSchemaPath = "fixture/invalid.json"
	MissingSchemaPath = "fixture/missing.json"
)

func TestCreatingAValidSchema(t *testing.T) {
	_, err := FromJSON(ValidSchemaPath)

	assert.Nil(t, err)
}

func TestCreatingAMissingSchema(t *testing.T) {
	_, err := FromJSON(MissingSchemaPath)

	assert.NotNil(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestCreatingAnInvalidSchema(t *testing.T) {
	_, err := FromJSON(InvalidSchemaPath)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "not-a-type is not a valid type")
}

func TestWrappingASchemaReader(t *testing.T) {
	s, err := FromJSON(ValidSchemaPath)
	require.Nil(t, err)

	sr := s.Reader(new(bytes.Buffer))
	wrapped := s.Reader(sr)

	assert.Equal(t, sr, wrapped)
}
