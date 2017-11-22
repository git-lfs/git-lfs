package tasklog

import "time"

// Task is an interface which encapsulates an activity which can be logged.
type Task interface {
	// Updates returns a channel which is written to with the current state
	// of the Task when an update is present. It is closed when the task is
	// complete.
	Updates() <-chan *Update

	// Throttled returns whether or not updates from this task should be
	// limited when being printed to a sink via *log.Logger.
	//
	// It is expected to return the same value for a given Task instance.
	Throttled() bool
}

// Update is a single message sent (S) from a Task at a given time (At).
type Update struct {
	// S is the message sent in this update.
	S string
	// At is the time that this update was sent.
	At time.Time

	// Force determines if this update should not be throttled.
	Force bool
}

// Throttled determines whether this update should be throttled, based on the
// given earliest time of the next update. The caller should determine how often
// updates should be throttled. An Update with Force=true is never throttled.
func (u *Update) Throttled(next time.Time) bool {
	return !(u.Force || u.At.After(next))
}
