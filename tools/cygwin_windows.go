//go:build windows
// +build windows

package tools

import (
	"bytes"

	"github.com/git-lfs/git-lfs/v3/subprocess"
	"github.com/git-lfs/git-lfs/v3/tr"
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
		panic(tr.Tr.Get("unknown enabled state for %v", c))
	}
}

var (
	cygwinState cygwinSupport
)

func isCygwin() bool {
	if cygwinState != cygwinStateUnknown {
		return cygwinState.Enabled()
	}

	cmd, err := subprocess.ExecCommand("uname")
	if err != nil {
		return false
	}
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
