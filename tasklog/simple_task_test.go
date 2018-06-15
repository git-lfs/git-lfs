package tasklog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleTaskLogLogsUpdates(t *testing.T) {
	task := NewSimpleTask()

	var updates []*Update

	go func() {
		for update := range task.Updates() {
			updates = append(updates, update)
		}
		task.OnComplete()
	}()

	task.Log("Hello, world")
	task.Complete()

	require.Len(t, updates, 1)
	assert.Equal(t, "Hello, world", updates[0].S)
}

func TestSimpleTaskLogfLogsFormattedUpdates(t *testing.T) {
	task := NewSimpleTask()

	var updates []*Update

	go func() {
		for update := range task.Updates() {
			updates = append(updates, update)
		}
		task.OnComplete()
	}()

	task.Logf("Hello, world (%d)", 3+4)
	task.Complete()

	require.Len(t, updates, 1)
	assert.Equal(t, "Hello, world (7)", updates[0].S)
}

func TestSimpleTaskCompleteClosesUpdates(t *testing.T) {
	task := NewSimpleTask()

	select {
	case <-task.Updates():
		t.Fatalf("tasklog: unexpected update from *SimpleTask")
	default:
	}

	go func() {
		<-task.Updates()
		task.OnComplete()
	}()

	task.Complete()

	if _, ok := <-task.Updates(); ok {
		t.Fatalf("tasklog: expected (*SimpleTask).Updates() to be closed")
	}
}

func TestSimpleTaskIsNotThrottled(t *testing.T) {
	task := NewSimpleTask()

	throttled := task.Throttled()

	assert.False(t, throttled,
		"tasklog: expected *SimpleTask not to be Throttle()-d")
}
