// +build windows

package subprocess

import (
	"os/exec"
	"syscall"
)

// ExecCommand is a small platform specific wrapper around os/exec.Command
func ExecCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Env = env
	return cmd
}
