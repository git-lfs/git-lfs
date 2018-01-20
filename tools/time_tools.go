package tools

import (
	"time"
)

// IsExpiredAtOrIn returns whether or not the result of calling TimeAtOrIn is
// "expired" within "until" units of time from now.
func IsExpiredAtOrIn(from time.Time, until time.Duration, at time.Time, in time.Duration) (time.Time, bool) {
	expiration := TimeAtOrIn(from, at, in)
	if expiration.IsZero() {
		return expiration, false
	}

	return expiration, expiration.Before(time.Now().Add(until))
}

// TimeAtOrIn returns either "at", or the "in" duration added to the current
// time. TimeAtOrIn prefers to add a duration rather than return the "at"
// parameter.
func TimeAtOrIn(from, at time.Time, in time.Duration) time.Time {
	if in == 0 {
		return at
	}
	return from.Add(in)
}
