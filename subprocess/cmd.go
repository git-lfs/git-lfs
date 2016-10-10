package subprocess

import (
	"os/exec"
	"sync/atomic"

	"github.com/github/git-lfs/errors"
)

var (
	ErrAlreadyRunning = errors.New("already running")
	ErrNotRunning     = errors.New("not running")
)

// Command is the defualt constructor for the *Command type. It takes a `name` and
// `args...` as does the `exec.Cmd` constructor, and applies the same
// semantics.
func Command(name string, args ...string) *Cmd {
	return &Cmd{
		Cmd: exec.Command(name, args...),
	}
}

// Cmd is a wrapper type for the *exec.Cmd type. It applies the same
// semantics, but changes the behavior of the error that it returns when running
// a command. If a command exits with a non-zero code, the output from stderr
// will be wrapped in the error in its entirety.
type Cmd struct {
	*exec.Cmd

	// running is used to keep track of whether or not the command has
	// already begun running, and thus, whether it is safe to re-wire the
	// command's stderr.
	//
	// The value of `running` may either be 0, or 1. It is set atomically
	// using the atomic.CompareAndSwapUnit32 method.
	running uint32
}

// Start has the identical semantics of `*exec.Cmd.Start`. It starts, but
// does not block on the execution of, the underlying command.
//
// If there was an
// error starting the command, or the Start function was called more than once,
// then an error will be returned.
func (c *Cmd) Start() error {
	if swp := atomic.CompareAndSwapUint32(&c.running, 0, 1); swp {
		return ErrAlreadyRunning
	}

	if c.Cmd.Stderr == nil {
		c.Cmd.Stderr = &prefixSuffixSaver{N: 32 << 10}
	}

	if err := c.Cmd.Start(); err != nil {
		return err
	}

	return nil
}

// Wait waits for the command to be executed, and returns an error if it exited
// in a dirty (non-zero) state. If there was such an error, the prefix and
// suffix of the contents of stderr will be wrapped in the returned error.
//
// If the command has not already been started, it will be started
// transparently, returning any errors that it encountered (if any).
func (c *Cmd) Wait() error {
	if atomic.LoadUint32(&c.running) != 1 {
		if err := c.Start(); err != nil {
			return err
		}
	}

	return c.wrapError(c.Cmd.Wait())
}

// wrapError wraps the given error if the error was an `*exec.ExitError` and
// Stderr was not changed.
func (c *Cmd) wrapError(err error) error {
	if _, ok := err.(*exec.ExitError); ok {
		if ps, ok := c.Stderr.(*prefixSuffixSaver); ok && ps != nil {
			err = errors.Wrap(err, string(ps.Bytes()))
		}
	}

	return err
}
