// +build windows

package tools

import (
	"bytes"
	"fmt"

	"github.com/git-lfs/git-lfs/subprocess"
)

type cygwinSupport byte

const (
	cygwinStateUnknown cygwinSupport = iota
	cygwinStateEnabled
	cygwinStateDisabled
)

func (c cygwinSupport) Enabled() bool {
	switch c {
	case cygwinStateEnabled:
		return true
	case cygwinStateDisabled:
		return false
	default:
		panic(fmt.Sprintf("unknown enabled state for %v", c))
	}
}

var (
	cygwinState cygwinSupport
)

func isCygwin() bool {
	if cygwinState != cygwinStateUnknown {
		return cygwinState.Enabled()
	}

	cmd := subprocess.ExecCommand("uname")
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	if bytes.Contains(out, []byte("CYGWIN")) || bytes.Contains(out, []byte("MSYS")) {
		cygwinState = cygwinStateEnabled
	} else {
		cygwinState = cygwinStateDisabled
	}

	return cygwinState.Enabled()
}
