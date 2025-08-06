//go:build testtools
// +build testtools

// A simple Git LFS pointer extension that translates lower case characters
// to upper case characters and vise versa. This is used in the Git LFS
// integration tests.

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

func main() {
	log := openLog()

	if len(os.Args) != 4 || (os.Args[1] != "clean" && os.Args[1] != "smudge") || os.Args[2] != "--" {
		logErrorAndExit(log, "invalid arguments: %s", strings.Join(os.Args, " "))
	}

	if log != nil {
		fmt.Fprintf(log, "%s: %s\n", os.Args[1], os.Args[3])
	}

	reader := bufio.NewReader(os.Stdin)
	var err error
	for {
		var r rune
		r, _, err = reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}

		if unicode.IsLower(r) {
			r = unicode.ToUpper(r)
		} else if unicode.IsUpper(r) {
			r = unicode.ToLower(r)
		}

		os.Stdout.WriteString(string(r))
	}

	if err != nil {
		logErrorAndExit(log, "unable to read stdin: %s", err)
	}

	if log != nil {
		log.Close()
	}
	os.Exit(0)
}

func openLog() *os.File {
	logPath := os.Getenv("LFSTEST_EXT_LOG")
	if logPath == "" {
		return nil
	}

	log, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logErrorAndExit(nil, "unable to open log %q: %s", logPath, err)
	}

	return log
}

func logErrorAndExit(log *os.File, format string, vals ...interface{}) {
	msg := fmt.Sprintf(format, vals...)
	fmt.Fprintln(os.Stderr, msg)

	if log != nil {
		fmt.Fprintln(log, msg)
		log.Close()
	}

	os.Exit(1)
}
