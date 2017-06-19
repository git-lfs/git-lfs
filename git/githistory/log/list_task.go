package log

import "fmt"

// ListTask is a Task implementation that logs all updates in a list where each
// entry is line-delimited.
//
// For example:
//   entry #1
//   entry #2
//   msg: ..., done
type ListTask struct {
	msg string
	ch  chan string
}

// NewListTask instantiates a new *ListTask instance with the given message.
func NewListTask(msg string) *ListTask {
	return &ListTask{
		msg: msg,
		ch:  make(chan string, 1),
	}
}

// Entry logs a line-delimited task entry.
func (l *ListTask) Entry(update string) {
	l.ch <- fmt.Sprintf("%s\n", update)
}

func (l *ListTask) Complete() {
	l.ch <- fmt.Sprintf("%s: ...", l.msg)
	close(l.ch)
}

// Throttled implements the Task.Throttled function and ensures that all log
// updates are printed to the sink.
func (l *ListTask) Throttled() bool { return false }

// Updates implements the Task.Updates function and returns a channel of updates
// to log to the sink.
func (l *ListTask) Updates() <-chan string { return l.ch }
