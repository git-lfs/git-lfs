package schema

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaReaderWithValidPayload(t *testing.T) {
	schema, err := FromJSON(ValidSchemaPath)
	require.Nil(t, err)

	r := schema.Reader(strings.NewReader("1"))
	io.Copy(ioutil.Discard, r)

	assert.Nil(t, r.ValidationErr())
}

func TestSchemaReaderWithInvalidPayload(t *testing.T) {
	schema, err := FromJSON(ValidSchemaPath)
	require.Nil(t, err)

	r := schema.Reader(strings.NewReader("-1"))
	io.Copy(ioutil.Discard, r)

	assert.NotNil(t, r.ValidationErr())
}

func TestSchemaReaderBeforeValidation(t *testing.T) {
	schema, err := FromJSON(ValidSchemaPath)
	require.Nil(t, err)

	r := schema.Reader(strings.NewReader("1"))

	assert.Equal(t, errValidationIncomplete, r.ValidationErr())
}

func TestSchemaReaderDuringValidation(t *testing.T) {
	schema, err := FromJSON(ValidSchemaPath)
	require.Nil(t, err)

	r := schema.Reader(strings.NewReader("12"))

	var b [1]byte
	n, err := r.Read(b[:])

	assert.Equal(t, 1, n)
	assert.Nil(t, err)

	assert.Equal(t, errValidationIncomplete, r.ValidationErr())
}
