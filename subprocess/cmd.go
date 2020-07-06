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

func (c *Cmd) Run() error {
	c.trace()
	return c.Cmd.Run()
}

func (c *Cmd) Start() error {
	c.trace()
	return c.Cmd.Start()
}

func (c *Cmd) Output() ([]byte, error) {
	c.trace()
	return c.Cmd.Output()
}

func (c *Cmd) CombinedOutput() ([]byte, error) {
	c.trace()
	return c.Cmd.CombinedOutput()
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

func (c *Cmd) trace() {
	if len(c.Args) > 0 {
		Trace(c.Args[0], c.Args[1:]...)
	} else {
		Trace(c.Path)
	}
}

func newCmd(cmd *exec.Cmd) *Cmd {
	wrapped := &Cmd{Cmd: cmd}
	return wrapped
}
