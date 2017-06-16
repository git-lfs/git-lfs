package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListTaskCallsDoneWhenComplete(t *testing.T) {
	task := NewListTask("example")
	task.Complete()

	select {
	case update, ok := <-task.Updates():
		assert.Equal(t, "example: ...", update)
		assert.True(t, ok,
			"git/githistory/log: expected Updates() to remain open")
	default:
		t.Fatal("git/githistory/log: expected update from *ListTask")
	}

	select {
	case update, ok := <-task.Updates():
		assert.False(t, ok,
			"git/githistory.log: unexpected *ListTask.Update(): %s", update)
	default:
		t.Fatal("git/githistory/log: expected *ListTask.Updates() to be closed")
	}
}

func TestListTaskWritesEntries(t *testing.T) {
	task := NewListTask("example")
	task.Entry("1")

	select {
	case update, ok := <-task.Updates():
		assert.True(t, ok,
			"git/githistory/log: expected ListTask.Updates() to remain open")
		assert.Equal(t, "1\n", update)
	default:
		t.Fatal("git/githistory/log: expected task.Updates() to have an update")
	}
}

func TestListTaskIsDurable(t *testing.T) {
	task := NewListTask("example")

	durable := task.Durable()

	assert.True(t, durable,
		"git/githistory/log: expected *ListTask to be Durable()")
}
