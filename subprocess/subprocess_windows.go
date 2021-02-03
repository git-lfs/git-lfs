// +build windows

package subprocess

import (
	"os/exec"
	"syscall"
)

// ExecCommand is a small platform specific wrapper around os/exec.Command
func ExecCommand(name string, arg ...string) *Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Path, _ = LookPath(name)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Env = fetchEnvironment()
	return newCmd(cmd)
}
