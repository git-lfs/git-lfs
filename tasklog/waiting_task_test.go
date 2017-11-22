package tasklog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWaitingTaskDisplaysWaitingStatus(t *testing.T) {
	task := NewWaitingTask("example")

	assert.Equal(t, "example: ...", (<-task.Updates()).S)
}

func TestWaitingTaskCallsDoneWhenComplete(t *testing.T) {
	task := NewWaitingTask("example")

	select {
	case v, ok := <-task.Updates():
		if ok {
			assert.Equal(t, "example: ...", v.S)
		} else {
			t.Fatal("expected channel to be open")
		}
	default:
	}

	task.Complete()

	if _, ok := <-task.Updates(); ok {
		t.Fatalf("expected channel to be closed")
	}
}

func TestWaitingTaskPanicsWithMultipleDoneCalls(t *testing.T) {
	task := NewWaitingTask("example")

	task.Complete()

	defer func() {
		if err := recover(); err == nil {
			t.Fatal("tasklog: expected panic()")
		} else {
			if s, ok := err.(error); ok {
				assert.Equal(t, "close of closed channel", s.Error())
			} else {
				t.Fatal("tasklog: expected panic() to implement error")
			}
		}
	}()

	task.Complete()
}

func TestWaitingTaskIsThrottled(t *testing.T) {
	task := NewWaitingTask("example")

	throttled := task.Throttled()

	assert.True(t, throttled,
		"tasklog: expected *WaitingTask to be Throttle()-d")
}
