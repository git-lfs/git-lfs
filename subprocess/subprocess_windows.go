//go:build windows
// +build windows

package subprocess

import (
	"os/exec"
	"syscall"
)

// ExecCommand is a small platform specific wrapper around os/exec.Command
func ExecCommand(name string, arg ...string) (*Cmd, error) {
	cmd := exec.Command(name, arg...)
	var err error
	cmd.Path, err = LookPath(name)
	if err != nil {
		return nil, err
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Env = fetchEnvironment()
	return newCmd(cmd), nil
}
