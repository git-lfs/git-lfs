package log

// Task is an interface which encapsulates an activity which can be logged.
type Task interface {
	// Updates returns a channel which is written to with the current state
	// of the Task when an update is present. It is closed when the task is
	// complete.
	Updates() <-chan string

	// Throttled returns whether or not updates from this task should be
	// limited when being printed to a sink via *log.Logger.
	//
	// It is expected to return the same value for a given Task instance.
	Throttled() bool
}
