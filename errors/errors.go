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
// called. The stack is accessed via ErrorStack(). Calling ErrorStack() on a
// regular Go error will return an empty byte slice.

import (
	"fmt"
	"runtime"

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

type errorWithCause interface {
	Error() string
	Cause() error
}

func parentOf(err error) error {
	if c, ok := err.(errorWithCause); ok {
		return c.Cause()
	}

	return nil
}

// IsFatalError indicates that the error is fatal and the process should exit
// immediately after handling the error.
func IsFatalError(err error) bool {
	if e, ok := err.(interface {
		Fatal() bool
	}); ok {
		return e.Fatal()
	}
	if parent := parentOf(err); parent != nil {
		return IsFatalError(parent)
	}
	return false
}

// IsNotImplementedError indicates the client attempted to use a feature the
// server has not implemented (e.g. the batch endpoint).
func IsNotImplementedError(err error) bool {
	if e, ok := err.(interface {
		NotImplemented() bool
	}); ok {
		return e.NotImplemented()
	}
	if parent := parentOf(err); parent != nil {
		return IsNotImplementedError(parent)
	}
	return false
}

// IsAuthError indicates the client provided a request with invalid or no
// authentication credentials when credentials are required (e.g. HTTP 401).
func IsAuthError(err error) bool {
	if e, ok := err.(interface {
		AuthError() bool
	}); ok {
		return e.AuthError()
	}
	if parent := parentOf(err); parent != nil {
		return IsAuthError(parent)
	}
	return false
}

// IsInvalidPointerError indicates an attempt to parse data that was not a
// valid pointer.
func IsInvalidPointerError(err error) bool {
	if e, ok := err.(interface {
		InvalidPointer() bool
	}); ok {
		return e.InvalidPointer()
	}
	if parent := parentOf(err); parent != nil {
		return IsInvalidPointerError(parent)
	}
	return false
}

// IsInvalidRepoError indicates an operation was attempted from outside a git
// repository.
func IsInvalidRepoError(err error) bool {
	if e, ok := err.(interface {
		InvalidRepo() bool
	}); ok {
		return e.InvalidRepo()
	}
	if parent := parentOf(err); parent != nil {
		return IsInvalidRepoError(parent)
	}
	return false
}

// IsSmudgeError indicates an error while smudging a files.
func IsSmudgeError(err error) bool {
	if e, ok := err.(interface {
		SmudgeError() bool
	}); ok {
		return e.SmudgeError()
	}
	if parent := parentOf(err); parent != nil {
		return IsSmudgeError(parent)
	}
	return false
}

// IsCleanPointerError indicates an error while cleaning a file.
func IsCleanPointerError(err error) bool {
	if e, ok := err.(interface {
		CleanPointerError() bool
	}); ok {
		return e.CleanPointerError()
	}
	if parent := parentOf(err); parent != nil {
		return IsCleanPointerError(parent)
	}
	return false
}

// IsNotAPointerError indicates the parsed data is not an LFS pointer.
func IsNotAPointerError(err error) bool {
	if e, ok := err.(interface {
		NotAPointerError() bool
	}); ok {
		return e.NotAPointerError()
	}
	if parent := parentOf(err); parent != nil {
		return IsNotAPointerError(parent)
	}
	return false
}

// IsBadPointerKeyError indicates that the parsed data has an invalid key.
func IsBadPointerKeyError(err error) bool {
	if e, ok := err.(interface {
		BadPointerKeyError() bool
	}); ok {
		return e.BadPointerKeyError()
	}

	if parent := parentOf(err); parent != nil {
		return IsBadPointerKeyError(parent)
	}
	return false
}

// If an error is abad pointer error of any type, returns NotAPointerError
func StandardizeBadPointerError(err error) error {
	if IsBadPointerKeyError(err) {
		badErr := err.(badPointerKeyError)
		if badErr.Expected == "version" {
			return NewNotAPointerError(err)
		}
	}
	return err
}

// IsDownloadDeclinedError indicates that the smudge operation should not download.
// TODO: I don't really like using errors to control that flow, it should be refactored.
func IsDownloadDeclinedError(err error) bool {
	if e, ok := err.(interface {
		DownloadDeclinedError() bool
	}); ok {
		return e.DownloadDeclinedError()
	}
	if parent := parentOf(err); parent != nil {
		return IsDownloadDeclinedError(parent)
	}
	return false
}

// IsRetriableError indicates the low level transfer had an error but the
// caller may retry the operation.
func IsRetriableError(err error) bool {
	if e, ok := err.(interface {
		RetriableError() bool
	}); ok {
		return e.RetriableError()
	}
	if parent := parentOf(err); parent != nil {
		return IsRetriableError(parent)
	}
	return false
}

func GetInnerError(err error) error {
	return parentOf(err)
}

// ErrorSetContext sets a value in the error's context. If the error has not
// been wrapped, it does nothing.
func ErrorSetContext(err error, key string, value interface{}) {
	if e, ok := err.(errorWrapper); ok {
		e.Set(key, value)
	}
}

// ErrorGetContext gets a value from the error's context. If the error has not
// been wrapped, it returns an empty string.
func ErrorGetContext(err error, key string) interface{} {
	if e, ok := err.(errorWrapper); ok {
		return e.Get(key)
	}
	return ""
}

// ErrorDelContext removes a value from the error's context. If the error has
// not been wrapped, it does nothing.
func ErrorDelContext(err error, key string) {
	if e, ok := err.(errorWrapper); ok {
		e.Del(key)
	}
}

// ErrorContext returns the context map for an error if it is a wrappedError.
// If it is not a wrappedError it will return an empty map.
func ErrorContext(err error) map[string]interface{} {
	if e, ok := err.(errorWrapper); ok {
		return e.Context()
	}
	return nil
}

type errorWrapper interface {
	errorWithCause

	Set(string, interface{})
	Get(string) interface{}
	Del(string)
	Context() map[string]interface{}
}

// wrappedError is the base error wrapper. It provides a Message string, a
// stack, and a context map around a regular Go error.
type wrappedError struct {
	errorWithCause
	context map[string]interface{}
}

// newWrappedError creates a wrappedError.
func newWrappedError(err error, message string) errorWrapper {
	if err == nil {
		err = errors.New("Error")
	}

	var errWithCause errorWithCause

	if len(message) > 0 {
		errWithCause = errors.Wrap(err, message).(errorWithCause)
	} else if ewc, ok := err.(errorWithCause); ok {
		errWithCause = ewc
	} else {
		errWithCause = errors.Wrap(err, "LFS").(errorWithCause)
	}

	return &wrappedError{
		context:        make(map[string]interface{}),
		errorWithCause: errWithCause,
	}
}

// Set sets the value for the key in the context.
func (e wrappedError) Set(key string, val interface{}) {
	e.context[key] = val
}

// Get gets the value for a key in the context.
func (e wrappedError) Get(key string) interface{} {
	return e.context[key]
}

// Del removes a key from the context.
func (e wrappedError) Del(key string) {
	delete(e.context, key)
}

// Context returns the underlying context.
func (e wrappedError) Context() map[string]interface{} {
	return e.context
}

// Definitions for IsFatalError()

type fatalError struct {
	errorWrapper
}

func (e fatalError) Fatal() bool {
	return true
}

func NewFatalError(err error) error {
	return fatalError{newWrappedError(err, "Fatal error")}
}

// Definitions for IsNotImplementedError()

type notImplementedError struct {
	errorWrapper
}

func (e notImplementedError) NotImplemented() bool {
	return true
}

func NewNotImplementedError(err error) error {
	return notImplementedError{newWrappedError(err, "Not implemented")}
}

// Definitions for IsAuthError()

type authError struct {
	errorWrapper
}

func (e authError) AuthError() bool {
	return true
}

func NewAuthError(err error) error {
	return authError{newWrappedError(err, "Authentication required")}
}

// Definitions for IsInvalidPointerError()

type invalidPointerError struct {
	errorWrapper
}

func (e invalidPointerError) InvalidPointer() bool {
	return true
}

func NewInvalidPointerError(err error) error {
	return invalidPointerError{newWrappedError(err, "Invalid pointer")}
}

// Definitions for IsInvalidRepoError()

type invalidRepoError struct {
	errorWrapper
}

func (e invalidRepoError) InvalidRepo() bool {
	return true
}

func NewInvalidRepoError(err error) error {
	return invalidRepoError{newWrappedError(err, "Not in a git repository")}
}

// Definitions for IsSmudgeError()

type smudgeError struct {
	errorWrapper
}

func (e smudgeError) SmudgeError() bool {
	return true
}

func NewSmudgeError(err error, oid, filename string) error {
	e := smudgeError{newWrappedError(err, "Smudge error")}
	ErrorSetContext(e, "OID", oid)
	ErrorSetContext(e, "FileName", filename)
	return e
}

// Definitions for IsCleanPointerError()

type cleanPointerError struct {
	errorWrapper
}

func (e cleanPointerError) CleanPointerError() bool {
	return true
}

func NewCleanPointerError(err error, pointer interface{}, bytes []byte) error {
	e := cleanPointerError{newWrappedError(err, "Clean pointer error")}
	ErrorSetContext(e, "pointer", pointer)
	ErrorSetContext(e, "bytes", bytes)
	return e
}

// Definitions for IsNotAPointerError()

type notAPointerError struct {
	errorWrapper
}

func (e notAPointerError) NotAPointerError() bool {
	return true
}

func NewNotAPointerError(err error) error {
	return notAPointerError{newWrappedError(err, "Pointer file error")}
}

type badPointerKeyError struct {
	Expected string
	Actual   string

	errorWrapper
}

func (e badPointerKeyError) BadPointerKeyError() bool {
	return true
}

func NewBadPointerKeyError(expected, actual string) error {
	err := fmt.Errorf("Error parsing LFS Pointer. Expected key %s, got %s", expected, actual)
	return badPointerKeyError{expected, actual, newWrappedError(err, "")}
}

// Definitions for IsDownloadDeclinedError()

type downloadDeclinedError struct {
	errorWrapper
}

func (e downloadDeclinedError) DownloadDeclinedError() bool {
	return true
}

func NewDownloadDeclinedError(err error) error {
	return downloadDeclinedError{newWrappedError(err, "File missing and download is not allowed")}
}

// Definitions for IsRetriableError()

type retriableError struct {
	errorWrapper
}

func (e retriableError) RetriableError() bool {
	return true
}

func NewRetriableError(err error) error {
	return retriableError{newWrappedError(err, "")}
}

// Stack returns a byte slice containing the runtime.Stack()
func Stack() []byte {
	stackBuf := make([]byte, 1024*1024)
	written := runtime.Stack(stackBuf, false)
	return stackBuf[:written]
}
