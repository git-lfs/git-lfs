package tests

import (
	"testing"
)

func TestHappyPath(t *testing.T) {
	run := Setup(t)
	defer run.Teardown()

	run.Git("lfs", "track", "*.dat")
	run.WriteFile("a.dat", []byte("a"))
	run.Git("add", "a.dat")
	run.Git("commit", "-m", "add a.dat")

	AssertString(t, "a", string(run.ReadFile("a.dat")))
}
