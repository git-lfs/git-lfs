package log

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type ChanTask chan string

func (e ChanTask) Updates() <-chan string { return e }

func (e ChanTask) Durable() Bool { return false }

type DurableChanTask chan string

func (e DurableChanTask) Updates() <-chan string { return e }

func (e DurableChanTask) Durable() bool { return true }

func TestLoggerLogsTasks(t *testing.T) {
	var buf bytes.Buffer

	task := make(chan string)
	go func() {
		task <- "first"
		task <- "second"
		close(task)
	}()

	l := NewLogger(&buf)
	l.throttle = 0
	l.widthFn = func() int { return 0 }
	l.enqueue(ChanTask(task))
	l.Close()

	assert.Equal(t, "first\rsecond\rsecond, done\n", buf.String())
}

func TestLoggerLogsMultipleTasksInOrder(t *testing.T) {
	var buf bytes.Buffer

	t1 := make(chan string)
	go func() {
		t1 <- "first"
		t1 <- "second"
		close(t1)
	}()
	t2 := make(chan string)
	go func() {
		t2 <- "third"
		t2 <- "fourth"
		close(t2)
	}()

	l := NewLogger(&buf)
	l.throttle = 0
	l.widthFn = func() int { return 0 }
	l.enqueue(ChanTask(t1), ChanTask(t2))
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
	t1, t2 := make(chan string), make(chan string)

	l.widthFn = func() int { return 0 }
	l.enqueue(ChanTask(t1))

	t1 <- "first"
	l.enqueue(ChanTask(t2))
	close(t1)
	t2 <- "second"
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

	t1 := make(chan string)
	go func() {
		t1 <- "first"                     // t = 0     ms, throttle was open
		time.Sleep(10 * time.Millisecond) // t = 10    ms, throttle is closed
		t1 <- "second"                    // t = 10+ε  ms, throttle is closed
		time.Sleep(10 * time.Millisecond) // t = 20    ms, throttle is open
		t1 <- "third"                     // t = 20+ε  ms, throttle was open
		close(t1)                         // t = 20+2ε ms, throttle is closed
	}()

	l := NewLogger(&buf)
	l.widthFn = func() int { return 0 }
	l.throttle = 15 * time.Millisecond

	l.enqueue(ChanTask(t1))
	l.Close()

	assert.Equal(t, strings.Join([]string{
		"first\r",
		"third\r",
		"third, done\n",
	}, ""), buf.String())
}

func TestLoggerThrottlesLastWrite(t *testing.T) {
	var buf bytes.Buffer

	t1 := make(chan string)
	go func() {
		t1 <- "first"                     // t = 0     ms, throttle was open
		time.Sleep(10 * time.Millisecond) // t = 10    ms, throttle is closed
		t1 <- "second"                    // t = 10+ε  ms, throttle is closed
		close(t1)                         // t = 10+2ε ms, throttle is closed
	}()

	l := NewLogger(&buf)
	l.widthFn = func() int { return 0 }
	l.throttle = 15 * time.Millisecond

	l.enqueue(ChanTask(t1))
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

	t1 := make(chan string)
	go func() {
		t1 <- "first"  // t = 0+ε  ms, throttle is open
		t1 <- "second" // t = 0+2ε ms, throttle is closed
		close(t1)      // t = 0+3ε ms, throttle is closed
	}()

	l.enqueue(DurableChanTask(t1))
	l.Close()

	assert.Equal(t, strings.Join([]string{
		"first\r",
		"second\r",
		"second, done\n",
	}, ""), buf.String())
}
