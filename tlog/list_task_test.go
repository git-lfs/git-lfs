package tlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListTaskCallsDoneWhenComplete(t *testing.T) {
	task := NewListTask("example")
	task.Complete()

	select {
	case update, ok := <-task.Updates():
		assert.Equal(t, "example: ...", update.S)
		assert.True(t, ok,
			"tlog: expected Updates() to remain open")
	default:
		t.Fatal("tlog: expected update from *ListTask")
	}

	select {
	case update, ok := <-task.Updates():
		assert.False(t, ok,
			"git/githistory.log: unexpected *ListTask.Update(): %s", update)
	default:
		t.Fatal("tlog: expected *ListTask.Updates() to be closed")
	}
}

func TestListTaskWritesEntries(t *testing.T) {
	task := NewListTask("example")
	task.Entry("1")

	select {
	case update, ok := <-task.Updates():
		assert.True(t, ok,
			"tlog: expected ListTask.Updates() to remain open")
		assert.Equal(t, "1\n", update.S)
	default:
		t.Fatal("tlog: expected task.Updates() to have an update")
	}
}

func TestListTaskIsNotThrottled(t *testing.T) {
	task := NewListTask("example")

	throttled := task.Throttled()

	assert.False(t, throttled,
		"tlog: expected *ListTask to be Throttle()-d")
}
