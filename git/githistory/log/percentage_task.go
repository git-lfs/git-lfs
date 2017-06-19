package log

import (
	"fmt"
	"sync/atomic"
)

// PercentageTask is a task that is performed against a known number of
// elements.
type PercentageTask struct {
	// msg is the task message.
	msg string
	// n is the number of elements whose work has been completed. It is
	// managed sync/atomic.
	n uint64
	// total is the total number of elements to execute work upon.
	total uint64
	// ch is a channel which is written to when the task state changes and
	// is closed when the task is completed.
	ch chan string
}

func NewPercentageTask(msg string, total uint64) *PercentageTask {
	p := &PercentageTask{
		msg:   msg,
		total: total,
		ch:    make(chan string, 1),
	}
	p.Count(0)

	return p
}

// Count indicates that work has been completed against "n" number of elements,
// marking the task as complete if the total "n" given to all invocations of
// this method is equal to total.
//
// Count returns the new total number of (atomically managed) elements that have
// been completed.
func (c *PercentageTask) Count(n uint64) (new uint64) {
	new = atomic.AddUint64(&c.n, n)

	percentage := 100 * float64(new) / float64(c.total)
	msg := fmt.Sprintf("%s: %3.f%% (%d/%d)",
		c.msg, percentage, new, c.total)

	select {
	case c.ch <- msg:
	default:
		// Use a non-blocking write, since it's unimportant that callers
		// receive all updates.
	}

	if new >= c.total {
		close(c.ch)
	}

	return new
}

// Updates implements Task.Updates and returns a channel which is written to
// when the state of this task changes, and closed when the task is completed.
// has been completed.
func (c *PercentageTask) Updates() <-chan string {
	return c.ch
}

// Throttled implements Task.Throttled and returns true, indicating that this
// task is throttled.
func (c *PercentageTask) Throttled() bool { return true }
