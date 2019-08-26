package tasklog

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	isatty "github.com/mattn/go-isatty"
	"github.com/olekukonko/ts"
)

const (
	DefaultLoggingThrottle = 200 * time.Millisecond
)

// Logger logs a series of tasks to an io.Writer, processing each task in order
// until completion .
type Logger struct {
	// sink is the writer to write to.
	sink io.Writer

	// widthFn is a function that returns the width of the terminal that
	// this logger is running within.
	widthFn func() int

	// tty is true if sink is connected to a terminal
	tty bool

	// forceProgress forces progress status even when stdout is not a tty
	forceProgress bool

	// throttle is the minimum amount of time that must pass between each
	// instant data is logged.
	throttle time.Duration

	// queue is the incoming, unbuffered queue of tasks to enqueue.
	queue chan Task
	// tasks is the set of tasks to process.
	tasks chan Task
	// wg is a WaitGroup that is incremented when new tasks are enqueued,
	// and decremented when tasks finish.
	wg *sync.WaitGroup
}

// Option is the type for
type Option func(*Logger)

// ForceProgress returns an options function that configures forced progress status
// on the logger.
func ForceProgress(v bool) Option {
	return func(l *Logger) {
		l.forceProgress = v
	}
}

// NewLogger returns a new *Logger instance that logs to "sink" and uses the
// current terminal width as the width of the line. Will log progress status if
// stdout is a terminal or if forceProgress is true
func NewLogger(sink io.Writer, options ...Option) *Logger {
	if sink == nil {
		sink = ioutil.Discard
	}

	l := &Logger{
		sink:     sink,
		throttle: DefaultLoggingThrottle,
		widthFn: func() int {
			size, err := ts.GetSize()
			if err != nil {
				return 80
			}
			return size.Col()
		},
		queue: make(chan Task),
		tasks: make(chan Task),
		wg:    new(sync.WaitGroup),
	}

	for _, option := range options {
		option(l)
	}

	l.tty = tty(sink)

	go l.consume()

	return l
}

type hasFd interface {
	Fd() uintptr
}

// tty returns true if the writer is connected to a tty
func tty(writer io.Writer) bool {
	if v, ok := writer.(hasFd); ok {
		return isatty.IsTerminal(v.Fd()) || isatty.IsCygwinTerminal(v.Fd())
	}
	return false
}

// Close closes the queue and does not allow new Tasks to be `enqueue()`'d. It
// waits until the currently running Task has completed.
func (l *Logger) Close() {
	if l == nil {
		return
	}

	close(l.queue)

	l.wg.Wait()
}

// Waitier creates and enqueues a new *WaitingTask.
func (l *Logger) Waiter(msg string) *WaitingTask {
	t := NewWaitingTask(msg)
	l.Enqueue(t)

	return t
}

// Percentage creates and enqueues a new *PercentageTask.
func (l *Logger) Percentage(msg string, total uint64) *PercentageTask {
	t := NewPercentageTask(msg, total)
	l.Enqueue(t)

	return t
}

// List creates and enqueues a new *ListTask.
func (l *Logger) List(msg string) *ListTask {
	t := NewListTask(msg)
	l.Enqueue(t)

	return t
}

// List creates and enqueues a new *SimpleTask.
func (l *Logger) Simple() *SimpleTask {
	t := NewSimpleTask()
	l.Enqueue(t)

	return t
}

// Enqueue enqueues the given Tasks "ts".
func (l *Logger) Enqueue(ts ...Task) {
	if l == nil {
		for _, t := range ts {
			if t == nil {
				// NOTE: Do not allow nil tasks which are unable
				// to be completed.
				continue
			}
			go func(t Task) {
				for range t.Updates() {
					// Discard all updates.
				}
			}(t)
		}
		return
	}

	l.wg.Add(len(ts))
	for _, t := range ts {
		if t == nil {
			// NOTE: See above.
			continue
		}
		l.queue <- t
	}
}

// consume creates a pseudo-infinte buffer between the incoming set of tasks and
// the queue of tasks to work on.
func (l *Logger) consume() {
	go func() {
		// Process the single next task in sequence until completion,
		// then consume the next task.
		for task := range l.tasks {
			l.logTask(task)
		}
	}()

	defer close(l.tasks)

	pending := make([]Task, 0)

	for {
		// If there is a pending task, "peek" it off of the set of
		// pending tasks.
		var next Task
		if len(pending) > 0 {
			next = pending[0]
		}

		if next == nil {
			// If there was no pending task, wait for either a)
			// l.queue to close, or b) a new task to be submitted.
			task, ok := <-l.queue
			if !ok {
				// If the queue is closed, no more new tasks may
				// be added.
				return
			}

			// Otherwise, add a new task to the set of tasks to
			// process immediately, since there is no current
			// buffer.
			l.tasks <- task
		} else {
			// If there is a pending task, wait for either a) a
			// write to process the task to become non-blocking, or
			// b) a new task to enter the queue.
			select {
			case task, ok := <-l.queue:
				if !ok {
					// If the queue is closed, no more tasks
					// may be added.
					return
				}
				// Otherwise, add the next task to the set of
				// pending, active tasks.
				pending = append(pending, task)
			case l.tasks <- next:
				// Or "pop" the peeked task off of the pending
				// set.
				pending = pending[1:]
			}
		}
	}
}

// logTask logs the set of updates from a given task to the sink, then logs a
// "done." message, and then marks the task as done.
//
// By default, the *Logger throttles log entry updates to once per the duration
// of time specified by `l.throttle time.Duration`.
//
// If the duration if 0, or the task is "durable" (by implementing
// github.com/git-lfs/git-lfs/tasklog#DurableTask), then all entries will be
// logged.
func (l *Logger) logTask(task Task) {
	defer l.wg.Done()

	logAll := !task.Throttled()
	var last time.Time

	var update *Update
	for update = range task.Updates() {
		if !tty(os.Stdout) && !l.forceProgress {
			continue
		}
		if logAll || l.throttle == 0 || !update.Throttled(last.Add(l.throttle)) {
			l.logLine(update.S)
			last = update.At
		}
	}

	if update != nil {
		// If a task sent no updates, the last recorded update will be
		// nil. Given this, only log a message when there was at least
		// (1) update.
		l.log(fmt.Sprintf("%s, done.\n", update.S))
	}

	if v, ok := task.(interface {
		// OnComplete is called after the Task "task" is closed, but
		// before new tasks are accepted.
		OnComplete()
	}); ok {
		// If the Task implements this interface, call it and block
		// before accepting new tasks.
		v.OnComplete()
	}
}

// logLine writes a complete line and moves the cursor to the beginning of the
// line.
//
// It returns the number of bytes "n" written to the sink and the error "err",
// if one was encountered.
func (l *Logger) logLine(str string) (n int, err error) {
	padding := strings.Repeat(" ", maxInt(0, l.widthFn()-len(str)))

	return l.log(str + padding + "\r")
}

// log writes a string verbatim to the sink.
//
// It returns the number of bytes "n" written to the sink and the error "err",
// if one was encountered.
func (l *Logger) log(str string) (n int, err error) {
	return fmt.Fprint(l.sink, str)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
