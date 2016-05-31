// +build !windows

package subprocess

import (
	"os/exec"
	"syscall"

	"github.com/kr/pty"
)

// NewTty creates a pseudo-TTY for a command and modifies it appropriately so
// the command thinks it's a real terminal
func NewTty(cmd *exec.Cmd) *Tty {
	tty := &Tty{}
	tty.cmd = cmd
	// Assign pty/tty so git thinks it's a real terminal
	tty.outpty, tty.outtty, _ = pty.Open()
	cmd.Stdin = tty.outtty
	cmd.Stdout = tty.outtty
	tty.errpty, tty.errtty, _ = pty.Open()
	cmd.Stderr = tty.errtty
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setctty = true
	cmd.SysProcAttr.Setsid = true

	return tty
}
