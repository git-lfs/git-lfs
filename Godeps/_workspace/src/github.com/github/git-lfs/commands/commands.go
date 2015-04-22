package commands

import (
	"bytes"
	"fmt"
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	Debugging    = false
	ErrorBuffer  = &bytes.Buffer{}
	ErrorWriter  = io.MultiWriter(os.Stderr, ErrorBuffer)
	OutputWriter = io.MultiWriter(os.Stdout, ErrorBuffer)
	RootCmd      = &cobra.Command{
		Use:   "git-lfs",
		Short: "Git LFS provides large file storage to Git.",
		Run: func(cmd *cobra.Command, args []string) {
			versionCommand(cmd, args)
			cmd.Usage()
		},
	}
)

// Error prints a formatted message to Stderr.  It also gets printed to the
// panic log if one is created for this command.
func Error(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	fmt.Fprintln(ErrorWriter, line)
}

// Print prints a formatted message to Stdout.  It also gets printed to the
// panic log if one is created for this command.
func Print(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	fmt.Fprintln(OutputWriter, line)
}

// Exit prints a formatted message and exits.
func Exit(format string, args ...interface{}) {
	Error(format, args...)
	os.Exit(2)
}

// Debug prints a formatted message if debugging is enabled.  The formatted
// message also shows up in the panic log, if created.
func Debug(format string, args ...interface{}) {
	if !Debugging {
		return
	}
	log.Printf(format, args...)
}

// LoggedError prints a formatted message to Stderr and writes a stack trace for
// the error to a log file without exiting.
func LoggedError(err error, format string, args ...interface{}) {
	Error(format, args...)
	file := handlePanic(err)

	if len(file) > 0 {
		fmt.Fprintf(os.Stderr, "\nErrors logged to %s.\nUse `git lfs logs last` to view the log.\n", file)
	}
}

// Panic prints a formatted message, and writes a stack trace for the error to
// a log file before exiting.
func Panic(err error, format string, args ...interface{}) {
	LoggedError(err, format, args...)
	os.Exit(2)
}

func Run() {
	RootCmd.Execute()
}

func PipeMediaCommand(name string, args ...string) error {
	return PipeCommand("bin/"+name, args...)
}

func PipeCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func handlePanic(err error) string {
	if err == nil {
		return ""
	}

	return logPanic(err, false)
}

func logEnv(w io.Writer) {
	for _, env := range lfs.Environ() {
		fmt.Fprintln(w, env)
	}
}

func logPanic(loggedError error, recursive bool) string {
	var fmtWriter io.Writer = os.Stderr

	if err := os.MkdirAll(lfs.LocalLogDir, 0755); err != nil {
		fmt.Fprintf(fmtWriter, "Unable to log panic to %s: %s\n\n", lfs.LocalLogDir, err.Error())
		return ""
	}

	now := time.Now()
	name := now.Format("20060102T150405.999999999")
	full := filepath.Join(lfs.LocalLogDir, name+".log")

	file, err := os.Create(full)
	if err == nil {
		fmtWriter = file
		defer file.Close()
	}

	fmt.Fprintf(fmtWriter, "> %s", filepath.Base(os.Args[0]))
	if len(os.Args) > 0 {
		fmt.Fprintf(fmtWriter, " %s", strings.Join(os.Args[1:], " "))
	}
	fmt.Fprint(fmtWriter, "\n")

	logEnv(fmtWriter)
	fmt.Fprint(fmtWriter, "\n")

	fmtWriter.Write(ErrorBuffer.Bytes())
	fmt.Fprint(fmtWriter, "\n")

	fmt.Fprintln(fmtWriter, loggedError.Error())

	if wErr, ok := loggedError.(ErrorWithStack); ok {
		fmt.Fprintln(fmtWriter, wErr.InnerError())
		for key, value := range wErr.Context() {
			fmt.Fprintf(fmtWriter, "%s=%s\n", key, value)
		}
		fmtWriter.Write(wErr.Stack())
	} else {
		fmtWriter.Write(lfs.Stack())
	}

	if err != nil && !recursive {
		fmt.Fprintf(fmtWriter, "Unable to log panic to %s\n\n", full)
		logPanic(err, true)
	}

	return full
}

type ErrorWithStack interface {
	Context() map[string]string
	InnerError() string
	Stack() []byte
}

func init() {
	log.SetOutput(ErrorWriter)
}
