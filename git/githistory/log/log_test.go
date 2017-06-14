package log

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ChanTask chan string

func (e ChanTask) Updates() <-chan string { return e }

func TestLoggerLogsTasks(t *testing.T) {
	var buf bytes.Buffer

	task := make(chan string)
	go func() {
		task <- "first"
		task <- "second"
		close(task)
	}()

	l := NewLogger(&buf)
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
