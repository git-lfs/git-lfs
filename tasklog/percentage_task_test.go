package tasklog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPercentageTaskCalculuatesPercentages(t *testing.T) {
	task := NewPercentageTask("example", 10)

	assert.Equal(t, "example:   0% (0/10)", (<-task.Updates()).S)

	n := task.Count(3)
	assert.EqualValues(t, 3, n)

	assert.Equal(t, "example:  30% (3/10)", (<-task.Updates()).S)
}

func TestPercentageTaskCalculatesPercentWithoutTotal(t *testing.T) {
	task := NewPercentageTask("example", 0)

	select {
	case v, ok := <-task.Updates():
		if ok {
			assert.Equal(t, "example: 100% (0/0)", v.S)
		} else {
			t.Fatal("expected channel to be open")
		}
	default:
	}
}

func TestPercentageTaskCallsDoneWhenComplete(t *testing.T) {
	task := NewPercentageTask("example", 10)

	select {
	case v, ok := <-task.Updates():
		if ok {
			assert.Equal(t, "example:   0% (0/10)", v.S)
		} else {
			t.Fatal("expected channel to be open")
		}
	default:
	}

	assert.EqualValues(t, 10, task.Count(10))
	assert.Equal(t, "example: 100% (10/10)", (<-task.Updates()).S)

	if _, ok := <-task.Updates(); ok {
		t.Fatalf("expected channel to be closed")
	}
}

func TestPercentageTaskIsThrottled(t *testing.T) {
	task := NewPercentageTask("example", 10)

	throttled := task.Throttled()

	assert.True(t, throttled,
		"tasklog: expected *PercentageTask to be Throttle()-d")
}

func TestPercentageTaskPanicsWhenOvercounted(t *testing.T) {
	task := NewPercentageTask("example", 0)
	defer func() {
		assert.Equal(t, "tasklog: counted too many items", recover())
	}()

	task.Count(1)
}
