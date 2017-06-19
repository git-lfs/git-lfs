package log

import "fmt"

// WaitingTask represents a task for which the total number of items to do work
// is on is unknown.
type WaitingTask struct {
	// ch is used to transmit task updates.
	ch chan string
}

// NewWaitingTask returns a new *WaitingTask.
func NewWaitingTask(msg string) *WaitingTask {
	ch := make(chan string, 1)
	ch <- fmt.Sprintf("%s: ...", msg)

	return &WaitingTask{ch: ch}
}

// Complete marks the task as completed.
func (w *WaitingTask) Complete() {
	close(w.ch)
}

// Done implements Task.Done and returns a channel which is closed when
// Complete() is called.
func (w *WaitingTask) Updates() <-chan string {
	return w.ch
}

// Durable implements Task.Durable and returns false, indicating that this task
// is not durable.
func (w *WaitingTask) Durable() bool { return false }
