package errors_test

import (
	"net/url"
	"testing"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/stretchr/testify/assert"
)

type TemporaryError struct {
}

func (e TemporaryError) Error() string {
	return ""
}

func (e TemporaryError) Temporary() bool {
	return true
}

type TimeoutError struct {
}

func (e TimeoutError) Error() string {
	return ""
}

func (e TimeoutError) Timeout() bool {
	return true
}

func TestCanRetryOnTemporaryError(t *testing.T) {
	err := &url.Error{Err: TemporaryError{}}
	assert.True(t, errors.IsRetriableError(err))
}

func TestCanRetryOnTimeoutError(t *testing.T) {
	err := &url.Error{Err: TimeoutError{}}
	assert.True(t, errors.IsRetriableError(err))
}

func TestCannotRetryOnGenericUrlError(t *testing.T) {
	err := &url.Error{Err: errors.New("")}
	assert.False(t, errors.IsRetriableError(err))
}
