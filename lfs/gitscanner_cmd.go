package lfs

import (
	"bufio"
	"io"
	"os/exec"
	"strings"

	"github.com/rubyist/tracerx"
)

type wrappedCmd struct {
	Stdin  io.WriteCloser
	Stdout *bufio.Reader
	Stderr *bufio.Reader
	*exec.Cmd
}

// startCommand starts up a command and creates a stdin pipe and a buffered
// stdout & stderr pipes, wrapped in a wrappedCmd. The stdout buffer will be of stdoutBufSize
// bytes.
func startCommand(command string, args ...string) (*wrappedCmd, error) {
	cmd := exec.Command(command, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	tracerx.Printf("run_command: %s %s", command, strings.Join(args, " "))
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &wrappedCmd{
		stdin,
		bufio.NewReaderSize(stdout, stdoutBufSize),
		bufio.NewReaderSize(stderr, stdoutBufSize),
		cmd,
	}, nil
}
