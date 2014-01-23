package gitmedia

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
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

// Debug prints a formatted message if debugging is enabled.  The formatted
// message also shows up in the panic log, if created.
func Debug(format string, args ...interface{}) {
	if !Debugging {
		return
	}
	log.Printf(format, args...)
}

func SetupDebugging(flagset *flag.FlagSet) {
	if flagset == nil {
		flag.BoolVar(&Debugging, "debug", false, "Turns debugging on")
	} else {
		flagset.BoolVar(&Debugging, "debug", false, "Turns debugging on")
	}
}

func handlePanic(err error) string {
	if err == nil {
		return ""
	}

	Debug(err.Error())
	logFile, logErr := logPanic(err)
	if logErr != nil {
		fmt.Fprintf(os.Stderr, "Unable to log panic to %s\n", LocalLogDir)
		logEnv(os.Stderr)
		panic(logErr)
	}

	return logFile
}

func logEnv(w io.Writer) {
	fmt.Fprintf(w, "TempDir=%s\n", TempDir)
	fmt.Fprintf(w, "LocalMediaDir=%s\n", LocalMediaDir)

	for _, env := range os.Environ() {
		if !strings.Contains(env, "GIT_") {
			continue
		}
		fmt.Fprintln(w, env)
	}
}

func logPanic(loggedError error) (string, error) {
	if err := os.MkdirAll(LocalLogDir, 0744); err != nil {
		return "", err
	}

	now := time.Now()
	name := now.Format("2006-01-02T15:04:05.999999999")
	full := filepath.Join(LocalLogDir, name+".log")

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
