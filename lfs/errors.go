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
	"fmt"
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

// IsAuthError indicates the client provided a request with invalid or no
// authentication credentials when credentials are required (e.g. HTTP 401).
func IsAuthError(err error) bool {
	if e, ok := err.(interface {
		AuthError() bool
	}); ok {
		return e.AuthError()
	}
	if e, ok := err.(errorWrapper); ok {
		return IsAuthError(e.InnerError())
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

// IsCleanPointerError indicates an error while cleaning a file.
func IsCleanPointerError(err error) bool {
	if e, ok := err.(interface {
		CleanPointerError() bool
	}); ok {
		return e.CleanPointerError()
	}
	if e, ok := err.(errorWrapper); ok {
		return IsCleanPointerError(e.InnerError())
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
	if e, ok := err.(errorWrapper); ok {
		return IsNotAPointerError(e.InnerError())
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
	if e, ok := err.(errorWrapper); ok {
		return IsBadPointerKeyError(e.InnerError())
	}
	return false
}

// IsDownloadDeclinedError indicates that the smudge operation should not download.
// TODO: I don't really like using errors to control that flow, it should be refactored.
func IsDownloadDeclinedError(err error) bool {
	if e, ok := err.(interface {
		DownloadDeclinedError() bool
	}); ok {
		return e.DownloadDeclinedError()
	}
	if e, ok := err.(errorWrapper); ok {
		return IsDownloadDeclinedError(e.InnerError())
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
	if e, ok := err.(errorWrapper); ok {
		return IsRetriableError(e.InnerError())
	}
	return false
}

func GetInnerError(err error) error {
	if e, ok := err.(interface {
		InnerError() error
	}); ok {
		return e.InnerError()
	}
	return nil
}

// Error wraps an error with an empty message.
func Error(err error) error {
	return Errorf(err, "")
}

// Errorf wraps an error with an additional formatted message.
func Errorf(err error, format string, args ...interface{}) error {
	if err == nil {
		err = errors.New("")
	}

	message := ""
	if len(format) > 0 {
		message = fmt.Sprintf(format, args...)
	}

	return newWrappedError(err, message)
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
func ErrorContext(err error) map[string]interface{} {
	if e, ok := err.(errorWrapper); ok {
		return e.Context()
	}
	return nil
}

type errorWrapper interface {
	InnerError() error
	Error() string
	Set(string, interface{})
	Get(string) interface{}
	Del(string)
	Context() map[string]interface{}
	Stack() []byte
}

// wrappedError is the base error wrapper. It provides a Message string, a
// stack, and a context map around a regular Go error.
type wrappedError struct {
	Message string
	stack   []byte
	context map[string]interface{}
	error
}

// newWrappedError creates a wrappedError. If the error has already been
// wrapped it is simply returned as is.
func newWrappedError(err error, message string) errorWrapper {
	if e, ok := err.(errorWrapper); ok {
		return e
	}

	if err == nil {
		err = errors.New("LFS Error")
	}

	if message == "" {
		message = err.Error()
	}

	return wrappedError{
		Message: message,
		stack:   Stack(),
		context: make(map[string]interface{}),
		error:   err,
	}
}

// Error will return the wrapped error's Message if it has one, otherwise it
// will call the underlying error's Error() function.
func (e wrappedError) Error() string {
	if e.Message == "" {
		return e.error.Error()
	}
	return e.Message
}

// InnerError returns the underlying error. This could be a Go error or another wrappedError.
func (e wrappedError) InnerError() error {
	return e.error
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
	return fatalError{newWrappedError(err, "Fatal error")}
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
	return notImplementedError{newWrappedError(err, "Not implemented")}
}

// Definitions for IsAuthError()

type authError struct {
	errorWrapper
}

func (e authError) InnerError() error {
	return e.errorWrapper
}

func (e authError) AuthError() bool {
	return true
}

func newAuthError(err error) error {
	return authError{newWrappedError(err, "Authentication required")}
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
	return invalidPointerError{newWrappedError(err, "Invalid pointer")}
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

func newInvalidRepoError(err error) error {
	return invalidRepoError{newWrappedError(err, "Not in a git repository")}
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
	e := smudgeError{newWrappedError(err, "Smudge error")}
	ErrorSetContext(e, "OID", oid)
	ErrorSetContext(e, "FileName", filename)
	return e
}

// Definitions for IsCleanPointerError()

type cleanPointerError struct {
	errorWrapper
}

func (e cleanPointerError) InnerError() error {
	return e.errorWrapper
}

func (e cleanPointerError) CleanPointerError() bool {
	return true
}

func newCleanPointerError(err error, pointer *Pointer, bytes []byte) error {
	e := cleanPointerError{newWrappedError(err, "Clean pointer error")}
	ErrorSetContext(e, "pointer", pointer)
	ErrorSetContext(e, "bytes", bytes)
	return e
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
	return notAPointerError{newWrappedError(err, "Not a valid Git LFS pointer file.")}
}

type badPointerKeyError struct {
	Expected string
	Actual   string
	errorWrapper
}

func (e badPointerKeyError) InnerError() error {
	return e.errorWrapper
}

func (e badPointerKeyError) BadPointerKeyError() bool {
	return true
}

func newBadPointerKeyError(expected, actual string) error {
	err := fmt.Errorf("Error parsing LFS Pointer. Expected key %s, got %s", expected, actual)
	return badPointerKeyError{expected, actual, newWrappedError(err, "")}
}

// Definitions for IsDownloadDeclinedError()

type downloadDeclinedError struct {
	errorWrapper
}

func (e downloadDeclinedError) InnerError() error {
	return e.errorWrapper
}

func (e downloadDeclinedError) DownloadDeclinedError() bool {
	return true
}

func newDownloadDeclinedError(err error) error {
	return downloadDeclinedError{newWrappedError(err, "File missing and download is not allowed")}
}

// Definitions for IsRetriableError()

type retriableError struct {
	errorWrapper
}

func (e retriableError) InnerError() error {
	return e.errorWrapper
}

func (e retriableError) RetriableError() bool {
	return true
}

func newRetriableError(err error) error {
	return retriableError{newWrappedError(err, "")}
}

// Stack returns a byte slice containing the runtime.Stack()
func Stack() []byte {
	stackBuf := make([]byte, 1024*1024)
	written := runtime.Stack(stackBuf, false)
	return stackBuf[:written]
}
