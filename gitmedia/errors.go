package gitmedia

import (
	"fmt"
	"runtime"
)

type WrappedError struct {
	Err     error
	Message string
	Stack   []byte
}

func Error(err error) *WrappedError {
	return Errorf(err, "")
}

func Errorf(err error, format string, args ...interface{}) *WrappedError {
	stackBuf := make([]byte, 1024*1024)
	written := runtime.Stack(stackBuf, false)

	e := &WrappedError{
		Err:     err,
		Message: err.Error(),
		Stack:   stackBuf[:written],
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
