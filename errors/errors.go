// Package errors provides common error handling tools
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package errors

// The LFS error system provides a simple wrapper around Go errors and the
// ability to inspect errors. It is strongly influenced by Dave Cheney's post
// at http://dave.cheney.net/2014/12/24/inspecting-errors.
//
// When passing errors out of lfs package functions, the return type should
// always be `error`. The wrappedError details are not exported. If an error is
// the kind of error a caller should need to investigate, an IsXError()
// function is provided that tells the caller if the error is of that type.
// There should only be a handfull of cases where a simple `error` is
// insufficient.
//
// The error behaviors can be nested when created. For example, the not
// implemented error can also be marked as a fatal error:
//
//	func LfsFunction() error {
//		err := functionCall()
//		if err != nil {
//			return newFatalError(newNotImplementedError(err))
//		}
//		return nil
//	}
//
// Then in the caller:
//
//	err := lfs.LfsFunction()
//	if lfs.IsNotImplementedError(err) {
//		log.Print("feature not implemented")
//	}
//	if lfs.IsFatalError(err) {
//		os.Exit(1)
//	}
//
// Wrapped errors contain a context, which is a map[string]string. These
// contexts can be accessed through the Error*Context functions. Calling these
// functions on a regular Go error will have no effect.
//
// Example:
//
//	err := lfs.SomeFunction()
//	errors.ErrorSetContext(err, "foo", "bar")
//	errors.ErrorGetContext(err, "foo") // => "bar"
//	errors.ErrorDelContext(err, "foo")
//
// Wrapped errors also contain the stack from the point at which they are
// called. Use the '%+v' printf verb to display. See the github.com/pkg/errors
// docs for more info: https://godoc.org/github.com/pkg/errors

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
)

// New returns an error with the supplied message. New also records the stack
// trace at thepoint it was called.
func New(message string) error {
	return errors.New(message)
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
// Errorf also records the stack trace at the point it was called.
func Errorf(format string, args ...interface{}) error {
	return errors.Errorf(format, args...)
}

// Wrap wraps an error with an additional message.
func Wrap(err error, msg string) error {
	return newWrappedError(err, msg)
}

// Wrapf wraps an error with an additional formatted message.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		err = errors.New("")
	}

	message := fmt.Sprintf(format, args...)

	return newWrappedError(err, message)
}

func StackTrace(err error) []string {
	type stacktrace interface {
		StackTrace() errors.StackTrace
	}

	if err, ok := err.(stacktrace); ok {
		frames := err.StackTrace()
		lines := make([]string, len(frames))
		for i, f := range frames {
			lines[i] = fmt.Sprintf("%+v", f)
		}
		return lines
	}

	return nil
}

func Combine(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for i, err := range errs {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(err.Error())
	}
	return fmt.Errorf(buf.String())
}

func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	if cause, ok := err.(causer); ok {
		return Cause(cause.Cause())
	}
	return err
}
