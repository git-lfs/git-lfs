package tasklog

import (
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpinnerTaskIsNotThrottled(t *testing.T) {
	task := NewSpinner()

	throttled := task.Throttled()

	assert.False(t, throttled,
		"tasklog: expected *SpinnerTask not to be Throttle()-d")
}

func TestSpinnerTaskFormatsMessages(t *testing.T) {
	task := NewSpinner()

	updates := make([]*Update, 0)
	wait := collectUpdates(task, &updates)

	task.Spinf("Hello %s.", "world")
	task.Finish("Goodbye.")

	wait.Wait()

	require.Len(t, updates, 2)
	assert.Equal(t, "| Hello world.", updates[0].S)
	assert.Equal(t, finishChar()+" Goodbye.", updates[1].S)
}

func TestSpinnerTaskUsesLastMessage(t *testing.T) {
	task := NewSpinner()

	updates := make([]*Update, 0)
	wait := collectUpdates(task, &updates)

	task.Spinf("Hello %s.", "world")
	task.Spin()
	task.Finish("Goodbye.")

	wait.Wait()

	require.Len(t, updates, 3)
	assert.Equal(t, "| Hello world.", updates[0].S)
	assert.Equal(t, "/ Hello world.", updates[1].S)
	assert.Equal(t, finishChar()+" Goodbye.", updates[2].S)
}

func collectUpdates(task Task, updates *[]*Update) *sync.WaitGroup {
	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		for u := range task.Updates() {
			*updates = append(*updates, u)
		}
		wg.Done()
	}()

	return wg
}

func finishChar() string {
	if runtime.GOOS == "windows" {
		return "*"
	}
	return "\u2714"
}
