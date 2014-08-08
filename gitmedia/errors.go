package gitmedia

import (
	"fmt"
	"runtime"
)

type WrappedError struct {
	Err     error
	Message string
	stack   []byte
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
		stack:   Stack(),
	}

	if len(format) > 0 {
		e.Errorf(format, args...)
	}

	return e
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

func Stack() []byte {
	stackBuf := make([]byte, 1024*1024)
	written := runtime.Stack(stackBuf, false)
	return stackBuf[:written]
}
