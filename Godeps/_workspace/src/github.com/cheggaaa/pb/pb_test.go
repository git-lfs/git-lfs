package pb

import (
	"fmt"
	"testing"
)

func Test_IncrementAddsOne(t *testing.T) {
	count := 5000
	bar := New(count)
	expected := 1
	actual := bar.Increment()

	if actual != expected {
		t.Error(fmt.Sprintf("Expected {%d} was {%d}", expected, actual))
	}
}
