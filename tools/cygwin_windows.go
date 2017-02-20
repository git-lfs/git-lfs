// +build windows

package tools

import (
	"bytes"

	"github.com/git-lfs/git-lfs/subprocess"
)

func isCygwin() bool {
	cmd := subprocess.ExecCommand("uname")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return bytes.Contains(out, []byte("CYGWIN"))
}
