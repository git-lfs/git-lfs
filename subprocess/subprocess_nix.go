// +build !windows

package subprocess

import (
	"os/exec"
)

// ExecCommand is a small platform specific wrapper around os/exec.Command
func ExecCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Env = env
	return cmd
}
