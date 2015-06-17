// +build !windows

// Package git contains various commands that shell out to git
package git

import (
	"os/exec"
)

// execCommand is a small platform specific wrapper around os/exec.Command
func execCommand(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}
