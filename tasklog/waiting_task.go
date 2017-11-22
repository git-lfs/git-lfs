package tasklog

import (
	"fmt"
	"time"
)

// WaitingTask represents a task for which the total number of items to do work
// is on is unknown.
type WaitingTask struct {
	// ch is used to transmit task updates.
	ch chan *Update
}

// NewWaitingTask returns a new *WaitingTask.
func NewWaitingTask(msg string) *WaitingTask {
	ch := make(chan *Update, 1)
	ch <- &Update{
		S:  fmt.Sprintf("%s: ...", msg),
		At: time.Now(),
	}

	return &WaitingTask{ch: ch}
}

// Complete marks the task as completed.
func (w *WaitingTask) Complete() {
	close(w.ch)
}

// Done implements Task.Done and returns a channel which is closed when
// Complete() is called.
func (w *WaitingTask) Updates() <-chan *Update {
	return w.ch
}

// Throttled implements Task.Throttled and returns true, indicating that this
// task is Throttled.
func (w *WaitingTask) Throttled() bool { return true }
