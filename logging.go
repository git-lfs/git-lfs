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
	Debugging   = false
	ErrorBuffer = &bytes.Buffer{}
	ErrorWriter = io.MultiWriter(os.Stderr, ErrorBuffer)
)

func Panic(err error, format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	fmt.Fprintln(ErrorWriter, line)

	if err != nil {
		Debug(err.Error())
		logErr := logPanic(err)
		if logErr != nil {
			fmt.Fprintln(os.Stderr, "Unable to log panic:")
			panic(logErr)
		}
	}
	os.Exit(2)
}

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

func logPanic(loggedError error) error {
	if err := os.MkdirAll(LocalLogDir, 0744); err != nil {
		return err
	}

	now := time.Now()
	name := now.Format("2006-01-02T15:04:05.999999999")
	full := filepath.Join(LocalLogDir, name+".log")

	file, err := os.Create(full)
	if err != nil {
		return err
	}

	defer file.Close()

	fmt.Fprintf(file, "> %s", filepath.Base(os.Args[0]))
	if len(os.Args) > 0 {
		fmt.Fprintf(file, " %s", strings.Join(os.Args[1:], " "))
	}
	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "")

	file.Write(ErrorBuffer.Bytes())
	fmt.Fprintln(file, "")

	fmt.Fprintln(file, loggedError.Error())
	file.Write(debug.Stack())

	return nil
}

func init() {
	log.SetOutput(ErrorWriter)
}
