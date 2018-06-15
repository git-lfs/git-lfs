package tasklog

import (
	"fmt"
	"sync"
	"time"
)

// SimpleTask is in an implementation of tasklog.Task which prints out messages
// verbatim.
type SimpleTask struct {
	// ch is used to transmit task updates.
	ch chan *Update

	// wg is used to wait between closing the channel, and acknowledging
	// that the close-related operations have been completed by the
	// tasklog.Logger.
	wg *sync.WaitGroup
}

// NewSimpleTask returns a new *SimpleTask instance.
func NewSimpleTask() *SimpleTask {
	return &SimpleTask{
		ch: make(chan *Update),
		wg: new(sync.WaitGroup),
	}
}

// Log logs a string with no formatting verbs.
func (s *SimpleTask) Log(str string) {
	s.Logf(str)
}

// Logf logs some formatted string, which is interpreted according to the rules
// defined in package "fmt".
func (s *SimpleTask) Logf(str string, vals ...interface{}) {
	s.ch <- &Update{
		S:  fmt.Sprintf(str, vals...),
		At: time.Now(),
	}
}

// Complete notes that the task is completed by closing the Updates channel and
// yields the logger to the next Task. Complete blocks until the *tasklog.Logger
// has acknowledged completion of this task.
func (s *SimpleTask) Complete() {
	s.wg.Add(1)
	close(s.ch)
	s.wg.Wait()
}

// OnComplete implements an interface which receives a call to this method when
// the *tasklog.Logger has finished processing this task, but before it has
// accepted new tasks.
func (s *SimpleTask) OnComplete() {
	s.wg.Done()
}

// Updates implements Task.Updates and returns a channel of updates which is
// closed when Complete() is called.
func (s *SimpleTask) Updates() <-chan *Update {
	return s.ch
}

// Throttled implements Task.Throttled and returns false, indicating that this
// task is not throttled.
func (s *SimpleTask) Throttled() bool { return false }
