package subprocess

import (
	"io"
	"os"
	"os/exec"
)

// Tty is a convenience wrapper to allow pseudo-TTYs on *nix systems, create with NewTty()
// Do not use any of the struct members directly, call the Stderr() and Stdout() methods
// Remember to call Close() when finished
type Tty struct {
	cmd    *exec.Cmd
	outpty *os.File
	outtty *os.File
	errpty *os.File
	errtty *os.File
}

func (t *Tty) Close() {
	if t.outtty != nil {
		t.outtty.Close()
		t.outtty = nil
	}
	if t.errtty != nil {
		t.errtty.Close()
		t.errtty = nil
	}
}

func (t *Tty) Stdout() (io.ReadCloser, error) {
	if t.outpty != nil {
		return t.outpty, nil
	} else {
		return t.cmd.StdoutPipe()
	}
}

func (t *Tty) Stderr() (io.ReadCloser, error) {
	if t.errpty != nil {
		return t.errpty, nil
	} else {
		return t.cmd.StderrPipe()
	}
}
