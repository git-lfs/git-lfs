package log

// Task is an interface which encapsulates an activity which can be logged.
type Task interface {
	// Updates returns a channel which is written to with the current state
	// of the Task when an update is present. It is closed when the task is
	// complete.
	Updates() <-chan string

	// Durable returns whether or not this task should be treated as
	// Durable.
	//
	// It is expected to return the same value for a given Task instance.
	Durable() bool
}
