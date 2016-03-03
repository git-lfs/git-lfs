// +build windows

// Package git contains various commands that shell out to git
package git

import (
	"os/exec"
	"syscall"
)

// execCommand is a small platform specific wrapper around os/exec.Command
func execCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Env = env
	return cmd
}
