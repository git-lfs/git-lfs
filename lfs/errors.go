package lfs

import (
	"fmt"
	"runtime"
)

type errorWrapper interface {
	InnerError() error
	Error() string
}

type wrappedError struct {
	Message string
	stack   []byte
	context map[string]string
	error
}

func newWrappedError(err error) errorWrapper {
	if e, ok := err.(errorWrapper); ok {
		return e
	}
	return wrappedError{
		Message: err.Error(),
		stack:   Stack(),
		context: make(map[string]string),
		error:   err,
	}
}

func (e wrappedError) InnerError() error {
	return e.error
}

func ErrorSetContext(err error, key, value string) {
	if e, ok := err.(wrappedError); ok {
		e.context[key] = value
	}
}

func ErrorGetContext(err error, key string) string {
	if e, ok := err.(wrappedError); ok {
		return e.context[key]
	}
	return ""
}

func ErrorDelContext(err error, key string) {
	if e, ok := err.(wrappedError); ok {
		delete(e.context, key)
	}
}

// ErrorStack returns the stack for an error if it is a wrappedError. If it is
// not a wrappedError it will return an empty byte slice.
func ErrorStack(err error) []byte {
	if e, ok := err.(wrappedError); ok {
		return e.stack
	}
	return nil
}

// ErrorContext returns the context map for an error if it is a wrappedError.
// If it is not a wrappedError it will return an empty map.
func ErrorContext(err error) map[string]string {
	if e, ok := err.(wrappedError); ok {
		return e.context
	}
	return nil
}

// fatalError indicates that the process should halt
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

func IsFatalError(err error) bool {
	type fatalerror interface {
		Fatal() bool
	}
	if e, ok := err.(fatalerror); ok {
		return e.Fatal()
	}
	if e, ok := err.(errorWrapper); ok {
		return IsFatalError(e.InnerError())
	}
	return false
}

// notImplementedError indicates that a feature (e.g. batch) is not implemented
// on the server.
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

func IsNotImplementedError(err error) bool {
	type notimplerror interface {
		NotImplemented() bool
	}
	if e, ok := err.(notimplerror); ok {
		return e.NotImplemented()
	}
	if e, ok := err.(errorWrapper); ok {
		return IsFatalError(e.InnerError())
	}
	return false
}

// invalidPointerError indicates the the pointer was invalid.
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

func IsInvalidPointerError(err error) bool {
	type invalidptrerror interface {
		InvalidPointer() bool
	}
	if e, ok := err.(invalidptrerror); ok {
		return e.InvalidPointer()
	}
	if e, ok := err.(errorWrapper); ok {
		return IsInvalidPointerError(e.InnerError())
	}
	return false
}

// invalidRepo error indicates that we are not in a git repository.
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

func IsInvalidRepoError(err error) bool {
	type invalidrepoerror interface {
		InvalidRepo() bool
	}
	if e, ok := err.(invalidrepoerror); ok {
		return e.InvalidRepo()
	}
	if e, ok := err.(errorWrapper); ok {
		return IsInvalidRepoError(e.InnerError())
	}
	return false
}

// Old

type WrappedError struct {
	Err     error
	Message string
	Panic   bool
	stack   []byte
	context map[string]string
}

func Error(err error) *WrappedError {
	return Errorf(err, "")
}

func Errorf(err error, format string, args ...interface{}) *WrappedError {
	if err == nil {
		return nil
	}

	e := &WrappedError{
		Err:     err,
		Message: err.Error(),
		Panic:   true,
		stack:   Stack(),
	}

	if len(format) > 0 {
		e.Errorf(format, args...)
	}

	return e
}

func (e *WrappedError) Set(key, value string) {
	if e.context == nil {
		e.context = make(map[string]string)
	}
	e.context[key] = value
}

func (e *WrappedError) Get(key string) string {
	if e.context == nil {
		return ""
	}
	return e.context[key]
}

func (e *WrappedError) Del(key string) {
	if e.context == nil {
		return
	}
	delete(e.context, key)
}

func (e *WrappedError) Error() string {
	return e.Message
}

func (e *WrappedError) Errorf(format string, args ...interface{}) {
	e.Message = fmt.Sprintf(format, args...)
}

func (e *WrappedError) InnerError() string {
	return e.Err.Error()
}

func (e *WrappedError) Stack() []byte {
	return e.stack
}

func (e *WrappedError) Context() map[string]string {
	if e.context == nil {
		e.context = make(map[string]string)
	}
	return e.context
}

func Stack() []byte {
	stackBuf := make([]byte, 1024*1024)
	written := runtime.Stack(stackBuf, false)
	return stackBuf[:written]
}
