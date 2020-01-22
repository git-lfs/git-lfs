/*
Simple library for retry mechanism

slightly inspired by [Try::Tiny::Retry](https://metacpan.org/pod/Try::Tiny::Retry)

SYNOPSIS

http get with retry:

	url := "http://example.com"
	var body []byte

	err := retry.Do(
		func() error {
			resp, err := http.Get(url)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			return nil
		},
	)

	fmt.Println(body)

[next examples](https://github.com/avast/retry-go/tree/master/examples)


SEE ALSO

* [giantswarm/retry-go](https://github.com/giantswarm/retry-go) - slightly complicated interface.

* [sethgrid/pester](https://github.com/sethgrid/pester) - only http retry for http calls with retries and backoff

* [cenkalti/backoff](https://github.com/cenkalti/backoff) - Go port of the exponential backoff algorithm from Google's HTTP Client Library for Java. Really complicated interface.

* [rafaeljesus/retry-go](https://github.com/rafaeljesus/retry-go) - looks good, slightly similar as this package, don't have 'simple' `Retry` method

* [matryer/try](https://github.com/matryer/try) - very popular package, nonintuitive interface (for me)

BREAKING CHANGES

1.0.2 -> 2.0.0

* argument of `retry.Delay` is final delay (no multiplication by `retry.Units` anymore)

* function `retry.Units` are removed

* [more about this breaking change](https://github.com/avast/retry-go/issues/7)


0.3.0 -> 1.0.0

* `retry.Retry` function are changed to `retry.Do` function

* `retry.RetryCustom` (OnRetry) and `retry.RetryCustomWithOpts` functions are now implement via functions produces Options (aka `retry.OnRetry`)


*/
package retry

import (
	"fmt"
	"strings"
	"time"
)

// Function signature of retryable function
type RetryableFunc func() error

func Do(retryableFunc RetryableFunc, opts ...Option) error {
	var n uint

	//default
	config := &Config{
		attempts:      10,
		delay:         100 * time.Millisecond,
		onRetry:       func(n uint, err error) {},
		retryIf:       IsRecoverable,
		delayType:     BackOffDelay,
		lastErrorOnly: false,
	}

	//apply opts
	for _, opt := range opts {
		opt(config)
	}

	errorLog := make(Error, config.attempts)

	for n < config.attempts {
		err := retryableFunc()

		if err != nil {
			errorLog[n] = unpackUnrecoverable(err)

			if !config.retryIf(err) {
				break
			}

			config.onRetry(n, err)

			// if this is last attempt - don't wait
			if n == config.attempts-1 {
				break
			}

			delayTime := config.delayType(n, config)
			time.Sleep(delayTime)
		} else {
			return nil
		}

		n++
	}

	if config.lastErrorOnly {
		return errorLog[n]
	}
	return errorLog
}

// Error type represents list of errors in retry
type Error []error

// Error method return string representation of Error
// It is an implementation of error interface
func (e Error) Error() string {
	logWithNumber := make([]string, lenWithoutNil(e))
	for i, l := range e {
		if l != nil {
			logWithNumber[i] = fmt.Sprintf("#%d: %s", i+1, l.Error())
		}
	}

	return fmt.Sprintf("All attempts fail:\n%s", strings.Join(logWithNumber, "\n"))
}

func lenWithoutNil(e Error) (count int) {
	for _, v := range e {
		if v != nil {
			count++
		}
	}

	return
}

// WrappedErrors returns the list of errors that this Error is wrapping.
// It is an implementation of the `errwrap.Wrapper` interface
// in package [errwrap](https://github.com/hashicorp/errwrap) so that
// `retry.Error` can be used with that library.
func (e Error) WrappedErrors() []error {
	return e
}

type unrecoverableError struct {
	error
}

// Unrecoverable wraps an error in `unrecoverableError` struct
func Unrecoverable(err error) error {
	return unrecoverableError{err}
}

// IsRecoverable checks if error is an instance of `unrecoverableError`
func IsRecoverable(err error) bool {
	_, isUnrecoverable := err.(unrecoverableError)
	return !isUnrecoverable
}

func unpackUnrecoverable(err error) error {
	if unrecoverable, isUnrecoverable := err.(unrecoverableError); isUnrecoverable {
		return unrecoverable.error
	}

	return err
}
