// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"fmt"
	"io"
	"os"
	"time"
)

var stdout io.Writer = os.Stdout

// This is the standard writer that prints to standard output.
type ConsoleLogWriter struct{}

// This creates a new ConsoleLogWriter
func NewConsoleLogWriter() ConsoleLogWriter {
	return ConsoleLogWriter{}
}

// This is the ConsoleLogWriter's output method.
func (w ConsoleLogWriter) LogWrite(rec *LogRecord) {
	timestr := rec.Created.Format(time.StampMicro)
	fmt.Fprint(stdout, "[", timestr, "] [", levelStrings[rec.Level], "] ", rec.Message, "\n")
}

// Close flushes the log. Probably don't need this any more.
func (w ConsoleLogWriter) Close() {
	os.Stdout.Sync()
}
