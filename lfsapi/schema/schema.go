package schema

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
)

// Schema holds a JSON schema to be used for validation against various
// payloads.
type Schema struct {
	// s is the internal handle on the implementation of the JSON schema
	// specification.
	s *gojsonschema.Schema
}

// FromJSON constructs a new `*Schema` instance from the JSON schema at
// `schemaPath` relative to the package this code was called from.
//
// If the file could not be accessed, or was unable to be parsed as a valid JSON
// schema, an appropriate error will be returned. Otherwise, the `*Schema` will
// be returned with a nil error.
func FromJSON(schemaPath string) (*Schema, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	dir = filepath.ToSlash(dir)
	schemaPath = filepath.Join(dir, schemaPath)

	if _, err := os.Stat(schemaPath); err != nil {
		return nil, err
	}

	schema, err := gojsonschema.NewSchema(gojsonschema.NewReferenceLoader(
		// Platform compatibility: use "/" separators always for file://
		fmt.Sprintf("file:///%s", filepath.ToSlash(schemaPath)),
	))
	if err != nil {
		return nil, err
	}

	return &Schema{schema}, nil
}

// Reader wraps the given `io.Reader`, "r" as a `*schema.Reader`, allowing the
// contents passed through the reader to be inspected as conforming to the JSON
// schema or not.
//
// If the reader "r" already _is_ a `*schema.Reader`, it will be returned as-is.
func (s *Schema) Reader(r io.Reader) *Reader {
	if sr, ok := r.(*Reader); ok {
		return sr
	}

	rdr := &Reader{
		buf:    new(bytes.Buffer),
		schema: s.s,
	}
	rdr.r = io.TeeReader(r, rdr.buf)

	return rdr
}
