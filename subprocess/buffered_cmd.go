package subprocess

import (
	"bufio"
	"io"
)

const (
	// stdoutBufSize is the size of the buffers given to a sub-process stdout
	stdoutBufSize = 16384
)

type BufferedCmd struct {
	*Cmd

	Stdin  io.WriteCloser
	Stdout *bufio.Reader
	Stderr *bufio.Reader
}
