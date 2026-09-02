//go:build !windows
// +build !windows

package subprocess

import (
	"os/exec"
)

// ExecCommand is a small platform specific wrapper around os/exec.Command
func ExecCommand(name string, arg ...string) (*Cmd, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(path, arg...)
	cmd.Args[0] = name
	cmd.Env = fetchEnvironment()
	return newCmd(cmd), nil
}
