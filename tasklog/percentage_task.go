package tasklog

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
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
	ch chan *Update
	// wg is used to wait between closing the channel, and acknowledging
	// that the close-related operations have been completed by the
	// tasklog.Logger.
	wg *sync.WaitGroup
}

func NewPercentageTask(msg string, total uint64) *PercentageTask {
	p := &PercentageTask{
		msg:   msg,
		total: total,
		ch:    make(chan *Update, 1),
		wg:    new(sync.WaitGroup),
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
		panic("tasklog: counted too many items")
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
		if c.total != 0 {
			c.wg.Add(1)
			defer c.wg.Wait()
		}

		close(c.ch)
	}

	return new
}

func (c *PercentageTask) OnComplete() {
	c.wg.Done()
}

// Entry logs a line-delimited task entry.
func (t *PercentageTask) Entry(update string) {
	t.ch <- &Update{
		S:     fmt.Sprintf("%s\n", update),
		At:    time.Now(),
		Force: true,
	}
}

// Updates implements Task.Updates and returns a channel which is written to
// when the state of this task changes, and closed when the task is completed.
// has been completed.
func (c *PercentageTask) Updates() <-chan *Update {
	return c.ch
}

// Throttled implements Task.Throttled and returns true, indicating that this
// task is throttled.
func (c *PercentageTask) Throttled() bool { return true }
