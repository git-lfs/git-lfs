package trace

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var tracerOut io.Writer

// Trace prints tracing output following the semantics of the GIT_TRACE environment
// variable. If GIT_TRACE is set to "1", "2", or "true", messages will be printed
// on stderr. If the variable is set to an integer value greater than 1 and lower than
// 10 it will be interpreted as an open file descriptor and will try to write messages
// into this file descriptor. Alternatively, if this variable is set to an absolute
// path, messages will be logged to that file.
func Trace(format string, args ...interface{}) {
	if tracerOut != nil {
		fmt.Fprintf(tracerOut, "trace media: "+format+"\n", args...)
	}
}

func init() {
	trace := os.Getenv("GIT_TRACE")
	if trace == "" || strings.ToLower(trace) == "false" {
		return
	}

	fd, err := strconv.Atoi(trace)

	if err != nil {
		if filepath.IsAbs(trace) {
			tracerOut, err = os.OpenFile(trace, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not open '%s' for tracing: %s\nDefaulting to tracing on stderr...\n", trace, err)
				tracerOut = os.Stderr
			}
		} else if strings.ToLower(trace) == "true" {
			tracerOut = os.Stderr
		} else {
			return
		}
	} else {
		switch fd {
		case 0:
			return
		case 1, 2:
			tracerOut = os.Stderr
		default:
			tracerOut = os.NewFile(uintptr(fd), "trace")
		}
	}
}
