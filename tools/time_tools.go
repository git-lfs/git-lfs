package tools

import "time"

// TimeAtOrIn returns either "at", or the "in" duration added to the current
// time. TimeAtOrIn prefers to add a duration rather than return the "at"
// parameter.
func TimeAtOrIn(at time.Time, in time.Duration) time.Time {
	if in == 0 {
		return at
	}
	return time.Now().Add(in)
}
