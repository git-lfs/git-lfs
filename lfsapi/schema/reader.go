package schema

import (
	"bytes"
	"io"
	"strings"
	"sync/atomic"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/xeipuuv/gojsonschema"
)

var (
	// errValidationIncomplete is an error returned when `ValidationErr()`
	// is called while the reader is still processing data.
	errValidationIncomplete = errors.New("lfsapi/schema: validation incomplete")
)

// state represents the set of valid states a `*Reader` (see below) can be in
type state uint32

const (
	// stateNotStarted means the `*Reader` has no processed any data
	stateNotStarted state = iota
	// stateProcessing means the `*Reader` has received a `Read()` call at
	// least once, but has not received an `io.EOF` yet.
	stateProcessing
	// stateProcessed means the `*Reader` has received a `Read()` call at
	// least once and has gotten an `io.EOF`, meaning there is no more data
	// to process.
	stateProcessed
)

type Reader struct {
	// r is the underlying io.Reader this one is wrapping.
	r io.Reader
	// buf is the buffer of data read from the underlying reader
	buf *bytes.Buffer
	// schema is the *gojsonschema.Schema to valid the buffer against
	schema *gojsonschema.Schema

	// state is the current state that this `*Reader` is in, and is updated
	// atomically through `atomic.SetUint32`, and etc.
	state uint32

	// result stores the result of the schema validation
	result *gojsonschema.Result
	// resultErr stores the (optional) error returned from the schema
	// validation
	resultErr error
}

var _ io.Reader = (*Reader)(nil)

// Read implements io.Reader.Read, and returns exactly the data received from
// the underlying reader.
//
// Read also sometimes advances state, as according to the valid instances of
// the `state` from above. If transitioning into the `stateProcessed` state, the
// schema will be validated.
func (r *Reader) Read(p []byte) (n int, err error) {
	atomic.CompareAndSwapUint32(&r.state, uint32(stateNotStarted), uint32(stateProcessing))

	n, err = r.r.Read(p)
	if err == io.EOF {
		got := gojsonschema.NewStringLoader(r.buf.String())
		r.result, r.resultErr = r.schema.Validate(got)

		atomic.CompareAndSwapUint32(&r.state, uint32(stateProcessing), uint32(stateProcessed))
	}

	return
}

// ValidationErr returns an error assosciated with validating the data. If
// there was an error performing the validation itself, that error will be
// returned with priority. If the validation has not started, or is incomplete,
// an appropriate error will be returned.
//
// Otherwise, if any validation errors were present, an error will be returned
// containing all of the validation errors. If the data passed validation, a
// value of 'nil' will be returned instead.
func (r *Reader) ValidationErr() error {
	if r.resultErr != nil {
		return r.resultErr
	} else {
		switch state(atomic.LoadUint32(&r.state)) {
		case stateNotStarted, stateProcessing:
			return errValidationIncomplete
		}
	}

	if r.result.Valid() {
		return nil
	}

	msg := "Validation errors:\n"
	for _, e := range r.result.Errors() {
		msg = strings.Join([]string{msg, e.Description()}, "\n")
	}

	return errors.New(msg)
}
