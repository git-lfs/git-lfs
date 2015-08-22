package lfs

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
//	lfs.ErrorSetContext(err, "foo", "bar")
//	lfs.ErrorGetContext(err, "foo") // => "bar"
//	lfs.ErrorDelContext(err, "foo")
//
// Wrapped errors also contain the stack from the point at which they are
// called. The stack is accessed via ErrorStack(). Calling ErrorStack() on a
// regular Go error will return an empty byte slice.

import (
	"errors"
	"runtime"
)

// IsFatalError indicates that the error is fatal and the process should exit
// immediately after handling the error.
func IsFatalError(err error) bool {
	if e, ok := err.(interface {
		Fatal() bool
	}); ok {
		return e.Fatal()
	}
	if e, ok := err.(errorWrapper); ok {
		return IsFatalError(e.InnerError())
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
	if e, ok := err.(errorWrapper); ok {
		return IsNotImplementedError(e.InnerError())
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
	if e, ok := err.(errorWrapper); ok {
		return IsInvalidPointerError(e.InnerError())
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
	if e, ok := err.(errorWrapper); ok {
		return IsInvalidRepoError(e.InnerError())
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
	if e, ok := err.(errorWrapper); ok {
		return IsSmudgeError(e.InnerError())
	}
	return false
}

// IsCleanPointerError indicates an error while cleaning a file. Because of the
// structure of the code, this check also returns a *cleanPointerError because
// that's how a Pointer and []byte were passed up to the caller. This is not
// very clean and should be refactored. The returned *cleanPointerError MUST
// NOT be accessed if the bool value is false.
func IsCleanPointerError(err error) (*cleanPointerError, bool) {
	if e, ok := err.(interface {
		CleanPointerError() bool
	}); ok {
		cpe := err.(cleanPointerError)
		return &cpe, e.CleanPointerError()
	}
	if e, ok := err.(errorWrapper); ok {
		return IsCleanPointerError(e.InnerError())
	}
	return nil, false
}

// IsNotAPointerError indicates the parsed data is not an LFS pointer.
func IsNotAPointerError(err error) bool {
	if e, ok := err.(interface {
		NotAPointerError() bool
	}); ok {
		return e.NotAPointerError()
	}
	if e, ok := err.(errorWrapper); ok {
		return IsNotAPointerError(e.InnerError())
	}
	return false
}

// Error wraps an error with an empty message.
func Error(err error) error {
	return Errorf(err, "")
}

// Errorf wraps an error with an additional formatted message.
func Errorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	e := newWrappedError(err)

	// ERRTODO this isn't right
	/*
		if len(format) > 0 {
			we := e.(wrappedError)
			we.Message = fmt.Sprintf(format, args...)
		}
	*/

	return e
}

// ErrorSetContext sets a value in the error's context. If the error has not
// been wrapped, it does nothing.
func ErrorSetContext(err error, key, value string) {
	if e, ok := err.(errorWrapper); ok {
		e.Set(key, value)
	}
}

// ErrorGetContext gets a value from the error's context. If the error has not
// been wrapped, it returns an empty string.
func ErrorGetContext(err error, key string) string {
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

// ErrorStack returns the stack for an error if it is a wrappedError. If it is
// not a wrappedError it will return an empty byte slice.
func ErrorStack(err error) []byte {
	if e, ok := err.(errorWrapper); ok {
		return e.Stack()
	}
	return nil
}

// ErrorContext returns the context map for an error if it is a wrappedError.
// If it is not a wrappedError it will return an empty map.
func ErrorContext(err error) map[string]string {
	if e, ok := err.(errorWrapper); ok {
		return e.Context()
	}
	return nil
}

type errorWrapper interface {
	InnerError() error
	Error() string
	Set(string, string)
	Get(string) string
	Del(string)
	Context() map[string]string
	Stack() []byte
}

// wrappedError is the base error wrapper. It provides a Message string, a
// stack, and a context map around a regular Go error.
type wrappedError struct {
	Message string
	stack   []byte
	context map[string]string
	error
}

// newWrappedError creates a wrappedError. If the error has already been
// wrapped it is simply returned as is.
func newWrappedError(err error) errorWrapper {
	if e, ok := err.(errorWrapper); ok {
		return e
	}

	if err == nil {
		err = errors.New("LFS Error")
	}

	return wrappedError{
		Message: err.Error(),
		stack:   Stack(),
		context: make(map[string]string),
		error:   err,
	}
}

// InnerError returns the underlying error. This could be a Go error or another wrappedError.
func (e wrappedError) InnerError() error {
	return e.error
}

// Set sets the value for the key in the context.
func (e wrappedError) Set(key, val string) {
	e.context[key] = val
}

// Get gets the value for a key in the context.
func (e wrappedError) Get(key string) string {
	return e.context[key]
}

// Del removes a key from the context.
func (e wrappedError) Del(key string) {
	delete(e.context, key)
}

// Context returns the underlying context.
func (e wrappedError) Context() map[string]string {
	return e.context
}

// Stack returns the stack.
func (e wrappedError) Stack() []byte {
	return e.stack
}

// Definitions for IsFatalError()

type fatalError struct {
	errorWrapper
}

func (e fatalError) InnerError() error {
	return e.errorWrapper
}

func (e fatalError) Fatal() bool {
	return true
}

func newFatalError(err error) error {
	return fatalError{newWrappedError(err)}
}

// Definitions for IsNotImplementedError()

type notImplementedError struct {
	errorWrapper
}

func (e notImplementedError) InnerError() error {
	return e.errorWrapper
}

func (e notImplementedError) NotImplemented() bool {
	return true
}

func newNotImplementedError(err error) error {
	return notImplementedError{newWrappedError(err)}
}

// Definitions for IsInvalidPointerError()

type invalidPointerError struct {
	errorWrapper
}

func (e invalidPointerError) InnerError() error {
	return e.errorWrapper
}

func (e invalidPointerError) InvalidPointer() bool {
	return true
}

func newInvalidPointerError(err error) error {
	return invalidPointerError{newWrappedError(err)}
}

// Definitions for IsInvalidRepoError()

type invalidRepoError struct {
	errorWrapper
}

func (e invalidRepoError) InnerError() error {
	return e.errorWrapper
}

func (e invalidRepoError) InvalidRepo() bool {
	return true
}

func newInvalidrepoError(err error) error {
	return invalidRepoError{newWrappedError(err)}
}

// Definitions for IsSmudgeError()

type smudgeError struct {
	errorWrapper
}

func (e smudgeError) InnerError() error {
	return e.errorWrapper
}

func (e smudgeError) SmudgeError() bool {
	return true
}

func newSmudgeError(err error, oid, filename string) error {
	e := smudgeError{newWrappedError(err)}
	ErrorSetContext(e, "OID", oid)
	ErrorSetContext(e, "FileName", filename)
	return e
}

// Definitions for IsCleanPointerError()

type cleanPointerError struct {
	pointer *Pointer
	bytes   []byte
	errorWrapper
}

func (e cleanPointerError) InnerError() error {
	return e.errorWrapper
}

func (e cleanPointerError) CleanPointerError() bool {
	return true
}

func (e cleanPointerError) Pointer() *Pointer {
	return e.pointer
}

func (e cleanPointerError) Bytes() []byte {
	return e.bytes
}

func newCleanPointerError(err error, pointer *Pointer, bytes []byte) error {
	return cleanPointerError{
		pointer,
		bytes,
		newWrappedError(err),
	}
}

// Definitions for IsNotAPointerError()

type notAPointerError struct {
	errorWrapper
}

func (e notAPointerError) InnerError() error {
	return e.errorWrapper
}

func (e notAPointerError) NotAPointerError() bool {
	return true
}

func newNotAPointerError(err error) error {
	return notAPointerError{newWrappedError(err)}
}

// Stack returns a byte slice containing the runtime.Stack()
func Stack() []byte {
	stackBuf := make([]byte, 1024*1024)
	written := runtime.Stack(stackBuf, false)
	return stackBuf[:written]
}
