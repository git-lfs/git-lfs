package tests

import (
	"testing"
)

func TestHappyPath(t *testing.T) {
	run := Setup(t)
	defer run.Teardown()

	AssertCommand(t,
		run.Git("lfs", "track", "*.dat"),
		"Tracking *.dat",
	)

	run.WriteFile("a.dat", []byte("a"))
	run.Git("add", "a.dat")

	AssertCommandContains(t,
		run.Git("commit", "-m", "add a.dat"),
		"master (root-commit)",
		"1 file changed",
		"create mode 100755 a.dat",
	)

	AssertString(t, "a", string(run.ReadFile("a.dat")))
	run.Git("push", "origin", "master")

	AssertPointerBlob(run,
		"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb", 1,
		"master", "a.dat",
	)
}
