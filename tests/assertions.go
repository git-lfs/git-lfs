package tests

import (
	"testing"
)

func AssertString(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Fatalf("Expected %s, got %s", expected, actual)
	}
}
