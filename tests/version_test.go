package tests

import (
	"testing"
)

func TestVersion(t *testing.T) {
	run := Setup(t)
	defer run.Teardown()

	userAgent := run.Git("lfs", "version")

	AssertCommandContains(t,
		run.Git("lfs", "version", "--comics"),
		userAgent,
		"Nothing may see Gah Lak Tus and survive",
	)

	run.Cd(".git")
	AssertCommand(t,
		run.Git("lfs", "version"),
		userAgent,
	)

	AssertCommandContains(t,
		run.Git("lfs", "version", "--comics"),
		userAgent,
		"Nothing may see Gah Lak Tus and survive",
	)

	run.MkdirP("subDir")
	run.Cd("subDir")

	AssertCommand(t,
		run.Git("lfs", "version"),
		userAgent,
	)

	AssertCommandContains(t,
		run.Git("lfs", "version", "--comics"),
		userAgent,
		"Nothing may see Gah Lak Tus and survive",
	)
}
