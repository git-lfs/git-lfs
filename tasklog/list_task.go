package tasklog

import (
	"fmt"
	"time"
)

// ListTask is a Task implementation that logs all updates in a list where each
// entry is line-delimited.
//
// For example:
//   entry #1
//   entry #2
//   msg: ..., done.
type ListTask struct {
	msg string
	ch  chan *Update
}

// NewListTask instantiates a new *ListTask instance with the given message.
func NewListTask(msg string) *ListTask {
	return &ListTask{
		msg: msg,
		ch:  make(chan *Update, 1),
	}
}

// Entry logs a line-delimited task entry.
func (l *ListTask) Entry(update string) {
	l.ch <- &Update{
		S:  fmt.Sprintf("%s\n", update),
		At: time.Now(),
	}
}

func (l *ListTask) Complete() {
	l.ch <- &Update{
		S:  fmt.Sprintf("%s: ...", l.msg),
		At: time.Now(),
	}
	close(l.ch)
}

// Throttled implements the Task.Throttled function and ensures that all log
// updates are printed to the sink.
func (l *ListTask) Throttled() bool { return false }

// Updates implements the Task.Updates function and returns a channel of updates
// to log to the sink.
func (l *ListTask) Updates() <-chan *Update { return l.ch }
