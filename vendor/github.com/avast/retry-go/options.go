package retry

import (
	"time"
)

// Function signature of retry if function
type RetryIfFunc func(error) bool

// Function signature of OnRetry function
// n = count of attempts
type OnRetryFunc func(n uint, err error)

type DelayTypeFunc func(n uint, config *Config) time.Duration

type Config struct {
	attempts      uint
	delay         time.Duration
	onRetry       OnRetryFunc
	retryIf       RetryIfFunc
	delayType     DelayTypeFunc
	lastErrorOnly bool
}

// Option represents an option for retry.
type Option func(*Config)

// return the direct last error that came from the retried function
// default is false (return wrapped errors with everything)
func LastErrorOnly(lastErrorOnly bool) Option {
	return func(c *Config) {
		c.lastErrorOnly = lastErrorOnly
	}
}

// Attempts set count of retry
// default is 10
func Attempts(attempts uint) Option {
	return func(c *Config) {
		c.attempts = attempts
	}
}

// Delay set delay between retry
// default is 100ms
func Delay(delay time.Duration) Option {
	return func(c *Config) {
		c.delay = delay
	}
}

// DelayType set type of the delay between retries
// default is BackOff
func DelayType(delayType DelayTypeFunc) Option {
	return func(c *Config) {
		c.delayType = delayType
	}
}

// BackOffDelay is a DelayType which increases delay between consecutive retries
func BackOffDelay(n uint, config *Config) time.Duration {
	return config.delay * (1 << n)
}

// FixedDelay is a DelayType which keeps delay the same through all iterations
func FixedDelay(_ uint, config *Config) time.Duration {
	return config.delay
}

// OnRetry function callback are called each retry
//
// log each retry example:
//
//	retry.Do(
//		func() error {
//			return errors.New("some error")
//		},
//		retry.OnRetry(func(n uint, err error) {
//			log.Printf("#%d: %s\n", n, err)
//		}),
//	)
func OnRetry(onRetry OnRetryFunc) Option {
	return func(c *Config) {
		c.onRetry = onRetry
	}
}

// RetryIf controls whether a retry should be attempted after an error
// (assuming there are any retry attempts remaining)
//
// skip retry if special error example:
//
//	retry.Do(
//		func() error {
//			return errors.New("special error")
//		},
//		retry.RetryIf(func(err error) bool {
//			if err.Error() == "special error" {
//				return false
//			}
//			return true
//		})
//	)
//
// By default RetryIf stops execution if the error is wrapped using `retry.Unrecoverable`,
// so above example may also be shortened to:
//
//	retry.Do(
//		func() error {
//			return retry.Unrecoverable(errors.New("special error"))
//		}
//	)
func RetryIf(retryIf RetryIfFunc) Option {
	return func(c *Config) {
		c.retryIf = retryIf
	}
}
