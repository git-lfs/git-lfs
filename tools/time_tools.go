package tools

import "time"

// IsExpiredAtOrIn returns whether or not the result of calling TimeAtOrIn is
// "expired" within "until" units of time from now.
func IsExpiredAtOrIn(now time.Time, until time.Duration, at time.Time, in time.Duration) (time.Time, bool) {
	expiration := TimeAtOrIn(now, at, in)
	if expiration.IsZero() {
		return expiration, false
	}

	return expiration, expiration.Before(now.Add(until))
}

// TimeAtOrIn returns either "at", or the "in" duration added to the current
// time. TimeAtOrIn prefers to add a duration rather than return the "at"
// parameter.
func TimeAtOrIn(now, at time.Time, in time.Duration) time.Time {
	if in == 0 {
		return at
	}
	return now.Add(in)
}
