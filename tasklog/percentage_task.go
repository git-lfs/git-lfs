package tasklog

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/git-lfs/git-lfs/v3/tr"
)

// PercentageTask is a task that is performed against a known number of
// elements.
type PercentageTask struct {
	// members managed via sync/atomic must be aligned at the top of this
	// structure (see: https://github.com/git-lfs/git-lfs/pull/2880).

	// n is the number of elements whose work has been completed. It is
	// managed sync/atomic.
	n uint64
	// total is the total number of elements to execute work upon.
	total uint64
	// msg is the task message.
	msg string
	// ch is a channel which is written to when the task state changes and
	// is closed when the task is completed.
	ch chan *Update
}

func NewPercentageTask(msg string, total uint64) *PercentageTask {
	p := &PercentageTask{
		msg:   msg,
		total: total,
		ch:    make(chan *Update, 1),
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
	if new = atomic.AddUint64(&c.n, n); new > c.total {
		panic(fmt.Sprintf("tasklog: %s", tr.Tr.Get("counted too many items")))
	}

	var percentage float64
	if c.total == 0 {
		percentage = 100
	} else {
		percentage = 100 * float64(new) / float64(c.total)
	}

	c.ch <- &Update{
		S: fmt.Sprintf("%s: %3.f%% (%d/%d)",
			c.msg, math.Floor(percentage), new, c.total),
		At: time.Now(),
	}

	if new >= c.total {
		close(c.ch)
	}

	return new
}

// Entry logs a line-delimited task entry.
func (c *PercentageTask) Entry(update string) {
	c.ch <- &Update{
		S:     fmt.Sprintf("%s\n", update),
		At:    time.Now(),
		Force: true,
	}
}

// Complete notes that the task is completed by setting the number of
// completed elements to the total number of elements, and if necessary
// closing the Updates channel, which yields the logger to the next Task.
func (c *PercentageTask) Complete() {
	if count := atomic.SwapUint64(&c.n, c.total); count < c.total {
		close(c.ch)
	}
}

// Updates implements Task.Updates and returns a channel which is written to
// when the state of this task changes, and closed when the task is completed.
func (c *PercentageTask) Updates() <-chan *Update {
	return c.ch
}

// Throttled implements Task.Throttled and returns true, indicating that this
// task is throttled.
func (c *PercentageTask) Throttled() bool { return true }
