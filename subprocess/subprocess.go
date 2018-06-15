// Package subprocess provides helper functions for forking new processes
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package subprocess

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rubyist/tracerx"
)

// BufferedExec starts up a command and creates a stdin pipe and a buffered
// stdout & stderr pipes, wrapped in a BufferedCmd. The stdout buffer will be
// of stdoutBufSize bytes.
func BufferedExec(name string, args ...string) (*BufferedCmd, error) {
	cmd := ExecCommand(name, args...)
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

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &BufferedCmd{
		cmd,
		stdin,
		bufio.NewReaderSize(stdout, stdoutBufSize),
		bufio.NewReaderSize(stderr, stdoutBufSize),
	}, nil
}

// SimpleExec is a small wrapper around os/exec.Command.
func SimpleExec(name string, args ...string) (string, error) {
	Trace(name, args...)
	return Output(ExecCommand(name, args...))
}

func Output(cmd *Cmd) (string, error) {
	//start copied from Go 1.6 exec.go
	captureErr := cmd.Stderr == nil
	if captureErr {
		cmd.Stderr = &prefixSuffixSaver{N: 32 << 10}
	}
	//end copied from Go 1.6 exec.go

	out, err := cmd.Output()

	if exitError, ok := err.(*exec.ExitError); ok {
		// TODO for min Go 1.6+, replace with ExitError.Stderr
		errorOutput := strings.TrimSpace(string(cmd.Stderr.(*prefixSuffixSaver).Bytes()))
		if errorOutput == "" {
			// some commands might write nothing to stderr but something to stdout in error-conditions, in which case, we'll use that
			// in the error string
			errorOutput = strings.TrimSpace(string(out))
		}

		ran := cmd.Path
		if len(cmd.Args) > 1 {
			ran = fmt.Sprintf("%s %s", cmd.Path, quotedArgs(cmd.Args[1:]))
		}
		formattedErr := fmt.Errorf("Error running %s: '%s' '%s'", ran, errorOutput, strings.TrimSpace(exitError.Error()))

		// return "" as output in error case, for callers that don't care about errors but rely on "" returned, in-case stdout != ""
		return "", formattedErr
	}

	return strings.Trim(string(out), " \n"), err
}

func Trace(name string, args ...string) {
	tracerx.Printf("exec: %s %s", name, quotedArgs(args))
}

func quotedArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}

	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = fmt.Sprintf("'%s'", arg)
	}
	return strings.Join(quoted, " ")
}

// An env for an exec.Command without GIT_TRACE
var env []string
var traceEnv = "GIT_TRACE="

func init() {
	realEnv := os.Environ()
	env = make([]string, 0, len(realEnv))

	for _, kv := range realEnv {
		if strings.HasPrefix(kv, traceEnv) {
			continue
		}
		env = append(env, kv)
	}
}

// remaining code in file copied from Go 1.6 (c4fa25f4fc8f4419d0b0707bcdae9199a745face) exec.go and can be removed if moving to Go 1.6 minimum.
// go 1.6 adds ExitError.Stderr with nice prefix/suffix trimming, which could replace cmd.Stderr above

//start copied from Go 1.6 exec.go
// prefixSuffixSaver is an io.Writer which retains the first N bytes
// and the last N bytes written to it. The Bytes() methods reconstructs
// it with a pretty error message.
type prefixSuffixSaver struct {
	N         int // max size of prefix or suffix
	prefix    []byte
	suffix    []byte // ring buffer once len(suffix) == N
	suffixOff int    // offset to write into suffix
	skipped   int64

	// TODO(bradfitz): we could keep one large []byte and use part of it for
	// the prefix, reserve space for the '... Omitting N bytes ...' message,
	// then the ring buffer suffix, and just rearrange the ring buffer
	// suffix when Bytes() is called, but it doesn't seem worth it for
	// now just for error messages. It's only ~64KB anyway.
}

func (w *prefixSuffixSaver) Write(p []byte) (n int, err error) {
	lenp := len(p)
	p = w.fill(&w.prefix, p)

	// Only keep the last w.N bytes of suffix data.
	if overage := len(p) - w.N; overage > 0 {
		p = p[overage:]
		w.skipped += int64(overage)
	}
	p = w.fill(&w.suffix, p)

	// w.suffix is full now if p is non-empty. Overwrite it in a circle.
	for len(p) > 0 { // 0, 1, or 2 iterations.
		n := copy(w.suffix[w.suffixOff:], p)
		p = p[n:]
		w.skipped += int64(n)
		w.suffixOff += n
		if w.suffixOff == w.N {
			w.suffixOff = 0
		}
	}
	return lenp, nil
}

// fill appends up to len(p) bytes of p to *dst, such that *dst does not
// grow larger than w.N. It returns the un-appended suffix of p.
func (w *prefixSuffixSaver) fill(dst *[]byte, p []byte) (pRemain []byte) {
	if remain := w.N - len(*dst); remain > 0 {
		add := minInt(len(p), remain)
		*dst = append(*dst, p[:add]...)
		p = p[add:]
	}
	return p
}

func (w *prefixSuffixSaver) Bytes() []byte {
	if w.suffix == nil {
		return w.prefix
	}
	if w.skipped == 0 {
		return append(w.prefix, w.suffix...)
	}
	var buf bytes.Buffer
	buf.Grow(len(w.prefix) + len(w.suffix) + 50)
	buf.Write(w.prefix)
	buf.WriteString("\n... omitting ")
	buf.WriteString(strconv.FormatInt(w.skipped, 10))
	buf.WriteString(" bytes ...\n")
	buf.Write(w.suffix[w.suffixOff:])
	buf.Write(w.suffix[:w.suffixOff])
	return buf.Bytes()
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

//end copied from Go 1.6 exec.go
