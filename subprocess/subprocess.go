// Package subprocess provides helper functions for forking new processes
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package subprocess

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rubyist/tracerx"
)

// SimpleExec is a small wrapper around os/exec.Command.
func SimpleExec(name string, args ...string) (string, error) {
	tracerx.Printf("run_command: '%s' %s", name, strings.Join(args, " "))
	cmd := ExecCommand(name, args...)

	output, err := cmd.Output()
	if exitError, ok := err.(*exec.ExitError); ok {
		errorOutput := strings.TrimSpace(string(exitError.Stderr))
		if errorOutput == "" {
			// some commands might write nothing to stderr but something to stdout in error-conditions, in which case, we'll use that
			// in the error string
			errorOutput = strings.TrimSpace(string(output))
		}
		formattedErr := fmt.Errorf("Error running %s %s: '%s' '%s'", name, args, errorOutput, strings.TrimSpace(exitError.Error()))

		// return "" as output in error case, for callers that don't care about errors but rely on "" returned, in-case stdout != ""
		return "", formattedErr
	}

	return strings.Trim(string(output), " \n"), err
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
