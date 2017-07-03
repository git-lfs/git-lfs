package subprocess

import "os/exec"

// NewTty creates a pseudo-TTY for a command and modifies it appropriately so
// the command thinks it's a real terminal
func NewTty(cmd *exec.Cmd) *Tty {
	// Nothing special for Windows at this time
	tty := &Tty{}
	tty.cmd = cmd
	return tty
}
