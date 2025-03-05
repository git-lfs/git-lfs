package tools

import "github.com/git-lfs/git-lfs/v3/errors"

// Interface for all types of wrapper around a channel of results and an error channel
// Implementors will expose a type-specific channel for results
// Call the Wait() function after processing the results channel to catch any errors
// that occurred during the async processing
type ChannelWrapper interface {
	// Call this after processing results channel to check for async errors
	Wait() error
}

// Base implementation of channel wrapper to just deal with errors
type BaseChannelWrapper struct {
	errorChan <-chan error
}

func (w *BaseChannelWrapper) Wait() error {
	var multiErr error
	for err := range w.errorChan {
		// Combine in case multiple errors
		multiErr = errors.Join(multiErr, err)
	}

	return multiErr
}

func NewBaseChannelWrapper(errChan <-chan error) *BaseChannelWrapper {
	return &BaseChannelWrapper{errorChan: errChan}
}
