package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

// SchemaValidator uses the gojsonschema library to validate the JSON encoding
// of Go objects against a pre-defined JSON schema.
type SchemaValidator struct {
	// Schema is the JSON schema to validate against.
	//
	// Subject is the instance of Go type that will be validated.
	Schema, Subject gojsonschema.JSONLoader
}

func NewSchemaValidator(t *testing.T, schemaName string, got interface{}) *SchemaValidator {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Platform compatibility: use "/" separators always for file://
	dir = filepath.ToSlash(dir)

	schema := gojsonschema.NewReferenceLoader(fmt.Sprintf(
		"file:///%s/schema/%s", dir, schemaName),
	)

	marshalled, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}

	subject := gojsonschema.NewStringLoader(string(marshalled))

	return &SchemaValidator{
		Schema:  schema,
		Subject: subject,
	}
}

// Validate validates a Go object against JSON schema in a testing environment.
// If the validation fails, then the test will fail after logging all of the
// validation errors experienced by the validator.
func Validate(t *testing.T, schemaName string, got interface{}) {
	NewSchemaValidator(t, schemaName, got).Assert(t)
}

// Refute ensures that a particular Go object does not validate the JSON schema
// given.
//
// If validation against the schema is successful, then the test will fail after
// logging.
func Refute(t *testing.T, schemaName string, got interface{}) {
	NewSchemaValidator(t, schemaName, got).Refute(t)
}

// Assert preforms the validation assertion against the given *testing.T.
func (v *SchemaValidator) Assert(t *testing.T) {
	if result, err := gojsonschema.Validate(v.Schema, v.Subject); err != nil {
		t.Fatal(err)
	} else if !result.Valid() {
		for _, err := range result.Errors() {
			t.Logf("Validation error: %s", err.Description())
		}
		t.Fail()
	}
}

// Refute refutes that the given subject will validate against a particular
// schema.
func (v *SchemaValidator) Refute(t *testing.T) {
	if result, err := gojsonschema.Validate(v.Schema, v.Subject); err != nil {
		t.Fatal(err)
	} else if result.Valid() {
		t.Fatal("api/schema: expected validation to fail, succeeded")
	}
}
