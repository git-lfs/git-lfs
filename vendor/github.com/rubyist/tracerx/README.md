# tracerx

Tracerx is a simple tracing package that logs messages depending on environment variables.
It is very much inspired by git's GIT_TRACE mechanism.

[![GoDoc](https://godoc.org/github.com/rubyist/tracerx?status.svg)](https://godoc.org/github.com/rubyist/tracerx)

## Installation

```
  go get github.com/rubyist/tracerx
```

## Example

```go
tracerx.DefaultKey = "FOO"
tracerx.Printf("send message")

tracerx.PrintfKey("BAR", "do a thing")

t := time.Now()
// Do some stuff
tracerx.PerformanceSince("BAR", "command x", t)
```

This example will send tracing output based on the environment variables `FOO_TRACE` and `BAR_TRACE`.

The values control where the tracing is output as follows:

```
unset, 0, or "false":   no output
1, 2:                   stderr
absolute path:          output will be written to the file
3 - 10:                 output will be written to that file descriptor
```

If an associated `BAR_TRACE_PERFORMANCE` is set to 1 or "true", the `PerformanceSince` line will
output timing information.

Keys can also be disabled. See the GoDoc for full API documentation.

## Bugs, Issues, Feedback

Right here on GitHub: [https://github.com/rubyist/tracerx](https://github.com/rubyist/tracerx)
