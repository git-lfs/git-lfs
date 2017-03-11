package subprocess

import (
	"io"
	"os/exec"
)

// Thin wrapper around exec.Cmd. Takes care of pipe shutdown by
// keeping an internal reference to any created pipes. Whenever
// Cmd.Wait() is called, all created pipes are closed.
type Cmd struct {
	*exec.Cmd

	pipes []io.Closer
}

func (c *Cmd) StdoutPipe() (io.ReadCloser, error) {
	stdout, err := c.Cmd.StdoutPipe()
	c.pipes = append(c.pipes, stdout)
	return stdout, err
}

func (c *Cmd) StderrPipe() (io.ReadCloser, error) {
	stderr, err := c.Cmd.StderrPipe()
	c.pipes = append(c.pipes, stderr)
	return stderr, err
}

func (c *Cmd) StdinPipe() (io.WriteCloser, error) {
	stdin, err := c.Cmd.StdinPipe()
	c.pipes = append(c.pipes, stdin)
	return stdin, err
}

func (c *Cmd) Wait() error {
	for _, pipe := range c.pipes {
		pipe.Close()
	}

	return c.Cmd.Wait()
}

func newCmd(cmd *exec.Cmd) *Cmd {
	wrapped := &Cmd{Cmd: cmd}
	return wrapped
}
