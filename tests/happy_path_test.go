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
	run.Git("add", ".gitattributes")

	AssertCommandContains(t,
		run.Git("commit", "-m", "add a.dat"),
		"master (root-commit)",
		"2 files changed",
		"create mode 100755 a.dat",
		"create mode 100644 .gitattributes",
	)

	AssertString(t, "a", string(run.ReadFile("a.dat")))

	AssertPointerBlob(run,
		"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb", 1,
		"master", "a.dat",
	)

	RefuteServerObject(run,
		"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
	)

	AssertCommandContains(t,
		run.Git("push", "origin", "master"),
		"(1 of 1 files) 1 B / 1 B  100.00 %",
		"* [new branch]      master -> master",
	)

	AssertServerObject(run,
		"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
		[]byte("a"))

	AssertCommandContains(t,
		run.CloneTo("clone"),
		"Cloning into 'clone'",
		"Downloading a.dat (1 B)",
	)

	AssertString(t, "a", string(run.ReadFile("a.dat")))

	AssertPointerBlob(run,
		"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb", 1,
		"master", "a.dat",
	)
}
