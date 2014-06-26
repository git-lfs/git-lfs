package commands

import (
	"bytes"
	"fmt"
	"github.com/github/git-media/gitmedia"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

var (
	Debugging    = false
	ErrorBuffer  = &bytes.Buffer{}
	ErrorWriter  = io.MultiWriter(os.Stderr, ErrorBuffer)
	OutputWriter = io.MultiWriter(os.Stdout, ErrorBuffer)
	RootCmd      = &cobra.Command{
		Use:   "git-media",
		Short: "Git Media provides large file support to Git.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Git Media, yo")
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

// Panic prints a formatted message, and writes a stack trace for the error to
// a log file before exiting.
func Panic(err error, format string, args ...interface{}) {
	Error(format, args...)
	file := handlePanic(err)

	if len(file) > 0 {
		fmt.Fprintf(os.Stderr, "\nErrors logged to %s.\nUse `git media logs last` to view the log.\n", file)
	}
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

	Debug(err.Error())
	logFile, logErr := logPanic(err)
	if logErr != nil {
		fmt.Fprintf(os.Stderr, "Unable to log panic to %s - %s\n\n", gitmedia.LocalLogDir, err)
		logEnv(os.Stderr)
	}

	return logFile
}

func logEnv(w io.Writer) {
	for _, env := range gitmedia.Environ() {
		fmt.Fprintln(w, env)
	}
}

func logPanic(loggedError error) (string, error) {
	if err := os.MkdirAll(gitmedia.LocalLogDir, 0744); err != nil {
		return "", err
	}

	now := time.Now()
	name := now.Format("2006-01-02T15:04:05.999999999")
	full := filepath.Join(gitmedia.LocalLogDir, name+".log")

	file, err := os.Create(full)
	if err != nil {
		return "", err
	}

	defer file.Close()

	fmt.Fprintf(file, "> %s", filepath.Base(os.Args[0]))
	if len(os.Args) > 0 {
		fmt.Fprintf(file, " %s", strings.Join(os.Args[1:], " "))
	}
	fmt.Fprint(file, "\n")

	logEnv(file)
	fmt.Fprint(file, "\n")

	file.Write(ErrorBuffer.Bytes())
	fmt.Fprint(file, "\n")

	fmt.Fprintln(file, loggedError.Error())
	file.Write(debug.Stack())

	return full, nil
}

func init() {
	log.SetOutput(ErrorWriter)
}
