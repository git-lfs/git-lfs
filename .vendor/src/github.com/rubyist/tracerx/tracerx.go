// Package tracerx implements a simple tracer function that uses environment
// variables to control the output. It is a generalized package inspired by
// git's GIT_TRACE mechanism.
//
// By default, tracerx will look for the TRACERX_TRACE environment variable.
// The default can by changed by setting the DefaultKey.
//
// The values control where the tracing is output as follows:
//     unset, 0, or "false":   no output
//     1, 2:                   stderr
//     absolute path:          output will be written to the file
//     3 - 10:                 output will be written to that file descriptor
//
// By default, messages will be prefixed with "trace: ". This prefix can be
// modified by setting Prefix.
//
package tracerx

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var (
	DefaultKey = "TRACERX"
	Prefix     = "trace: "
	tracers    map[string]*tracer
	tracerLock sync.Mutex
)

type tracer struct {
	enabled bool
	w       io.Writer
}

// Printf writes a trace message for the DefaultKey
func Printf(format string, args ...interface{}) {
	PrintfKey(DefaultKey, format, args...)
}

// PrintfKey writes a trace message for the given key
func PrintfKey(key, format string, args ...interface{}) {
	uppedKey := strings.ToUpper(key)

	tracer, ok := tracers[uppedKey]
	if !ok {
		tracer = initializeTracer(uppedKey)
	}

	if tracer.enabled {
		fmt.Fprintf(tracer.w, Prefix+format+"\n", args...)
		return
	}
}

// Disable will disable tracing for the given key, regardless of
// the environment variable
func Disable(key string) {
	uppedKey := strings.ToUpper(key)
	if tracer, ok := tracers[uppedKey]; ok {
		tracer.enabled = false
	}
}

// Enable will enable tracing for the given key, regardless of
// the environment variable
func Enable(key string) {
	uppedKey := strings.ToUpper(key)
	if tracer, ok := tracers[uppedKey]; ok {
		tracer.enabled = true
	}
}

func initializeTracer(key string) *tracer {
	tracerLock.Lock()
	defer tracerLock.Unlock()

	if tracer, ok := tracers[key]; ok {
		return tracer
	}

	tracer := &tracer{false, os.Stderr}
	tracers[key] = tracer

	trace := os.Getenv(fmt.Sprintf("%s_TRACE", key))
	if trace == "" || strings.ToLower(trace) == "false" {
		return tracer
	}

	fd, err := strconv.Atoi(trace)
	if err != nil {
		// Not a number, it could be a path for a log file
		if filepath.IsAbs(trace) {
			tracerOut, err := os.OpenFile(trace, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not open '%s' for tracing: %s\nDefaulting to tracing on stderr...\n", trace, err)
				tracerOut = os.Stderr
			}
			tracer.w = tracerOut
			tracer.enabled = true
		} else if strings.ToLower(trace) == "true" {
			tracer.enabled = true
		}
	} else {
		switch fd {
		case 0:
		case 1, 2:
			tracer.enabled = true
		default:
			tracer.w = os.NewFile(uintptr(fd), "trace")
			tracer.enabled = true
		}
	}

	return tracer
}

func init() {
	tracers = make(map[string]*tracer, 0)
}
