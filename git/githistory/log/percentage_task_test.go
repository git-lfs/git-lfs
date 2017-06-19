package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPercentageTaskCalculuatesPercentages(t *testing.T) {
	task := NewPercentageTask("example", 10)

	assert.Equal(t, "example:   0% (0/10)", <-task.Updates())

	n := task.Count(3)
	assert.EqualValues(t, 3, n)

	assert.Equal(t, "example:  30% (3/10)", <-task.Updates())
}

func TestPercentageTaskCallsDoneWhenComplete(t *testing.T) {
	task := NewPercentageTask("example", 10)

	select {
	case v, ok := <-task.Updates():
		if ok {
			assert.Equal(t, "example:   0% (0/10)", v)
		} else {
			t.Fatal("expected channel to be open")
		}
	default:
	}

	assert.EqualValues(t, 10, task.Count(10))
	assert.Equal(t, "example: 100% (10/10)", <-task.Updates())

	if _, ok := <-task.Updates(); ok {
		t.Fatalf("expected channel to be closed")
	}
}
