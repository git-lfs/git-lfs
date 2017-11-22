package tasklog

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type ChanTask chan *Update

func (e ChanTask) Updates() <-chan *Update { return e }

func (e ChanTask) Throttled() bool { return true }

type UnthrottledChanTask chan *Update

func (e UnthrottledChanTask) Updates() <-chan *Update { return e }

func (e UnthrottledChanTask) Throttled() bool { return false }

func TestLoggerLogsTasks(t *testing.T) {
	var buf bytes.Buffer

	task := make(chan *Update)
	go func() {
		task <- &Update{"first", time.Now(), false}
		task <- &Update{"second", time.Now(), false}
		close(task)
	}()

	l := NewLogger(&buf)
	l.throttle = 0
	l.widthFn = func() int { return 0 }
	l.Enqueue(ChanTask(task))
	l.Close()

	assert.Equal(t, "first\rsecond\rsecond, done\n", buf.String())
}

func TestLoggerLogsMultipleTasksInOrder(t *testing.T) {
	var buf bytes.Buffer

	t1 := make(chan *Update)
	go func() {
		t1 <- &Update{"first", time.Now(), false}
		t1 <- &Update{"second", time.Now(), false}
		close(t1)
	}()
	t2 := make(chan *Update)
	go func() {
		t2 <- &Update{"third", time.Now(), false}
		t2 <- &Update{"fourth", time.Now(), false}
		close(t2)
	}()

	l := NewLogger(&buf)
	l.throttle = 0
	l.widthFn = func() int { return 0 }
	l.Enqueue(ChanTask(t1), ChanTask(t2))
	l.Close()

	assert.Equal(t, strings.Join([]string{
		"first\r",
		"second\r",
		"second, done\n",
		"third\r",
		"fourth\r",
		"fourth, done\n",
	}, ""), buf.String())
}

func TestLoggerLogsMultipleTasksWithoutBlocking(t *testing.T) {
	var buf bytes.Buffer

	l := NewLogger(&buf)
	l.throttle = 0
	t1, t2 := make(chan *Update), make(chan *Update)

	l.widthFn = func() int { return 0 }
	l.Enqueue(ChanTask(t1))

	t1 <- &Update{"first", time.Now(), false}
	l.Enqueue(ChanTask(t2))
	close(t1)
	t2 <- &Update{"second", time.Now(), false}
	close(t2)

	l.Close()

	assert.Equal(t, strings.Join([]string{
		"first\r",
		"first, done\n",
		"second\r",
		"second, done\n",
	}, ""), buf.String())
}

func TestLoggerThrottlesWrites(t *testing.T) {
	var buf bytes.Buffer

	t1 := make(chan *Update)
	go func() {
		start := time.Now()

		t1 <- &Update{"first", start, false}                             // t = 0     ms, throttle was open
		t1 <- &Update{"forced", start.Add(10 * time.Millisecond), true}  // t = 10+ε  ms, throttle is closed
		t1 <- &Update{"second", start.Add(10 * time.Millisecond), false} // t = 10+ε  ms, throttle is closed
		t1 <- &Update{"third", start.Add(26 * time.Millisecond), false}  // t = 20+ε  ms, throttle was open
		close(t1)                                                        // t = 20+2ε ms, throttle is closed
	}()

	l := NewLogger(&buf)
	l.widthFn = func() int { return 0 }
	l.throttle = 15 * time.Millisecond

	l.Enqueue(ChanTask(t1))
	l.Close()

	assert.Equal(t, strings.Join([]string{
		"first\r",
		"forced\r",
		"third\r",
		"third, done\n",
	}, ""), buf.String())
}

func TestLoggerThrottlesLastWrite(t *testing.T) {
	var buf bytes.Buffer

	t1 := make(chan *Update)
	go func() {
		start := time.Now()

		t1 <- &Update{"first", start, false}                             // t = 0     ms, throttle was open
		t1 <- &Update{"second", start.Add(10 * time.Millisecond), false} // t = 10+ε  ms, throttle is closed
		close(t1)                                                        // t = 10+2ε ms, throttle is closed
	}()

	l := NewLogger(&buf)
	l.widthFn = func() int { return 0 }
	l.throttle = 15 * time.Millisecond

	l.Enqueue(ChanTask(t1))
	l.Close()

	assert.Equal(t, strings.Join([]string{
		"first\r",
		"second, done\n",
	}, ""), buf.String())
}

func TestLoggerLogsAllDurableUpdates(t *testing.T) {
	var buf bytes.Buffer

	l := NewLogger(&buf)
	l.widthFn = func() int { return 0 }
	l.throttle = 15 * time.Minute

	t1 := make(chan *Update)
	go func() {
		t1 <- &Update{"first", time.Now(), false}  // t = 0+ε  ms, throttle is open
		t1 <- &Update{"second", time.Now(), false} // t = 0+2ε ms, throttle is closed
		close(t1)                                  // t = 0+3ε ms, throttle is closed
	}()

	l.Enqueue(UnthrottledChanTask(t1))
	l.Close()

	assert.Equal(t, strings.Join([]string{
		"first\r",
		"second\r",
		"second, done\n",
	}, ""), buf.String())
}

func TestLoggerHandlesSilentTasks(t *testing.T) {
	var buf bytes.Buffer

	task := make(chan *Update)
	close(task)

	l := NewLogger(&buf)
	l.Enqueue(ChanTask(task))
	l.Close()

	assert.Equal(t, "", buf.String())
}
