package subprocess

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

// SimpleExec is a small wrapper around os/exec.Command.
func SimpleExec(name string, args ...string) (string, error) {
	tracerx.Printf("run_command: '%s' %s", name, strings.Join(args, " "))
	cmd := ExecCommand(name, args...)

	output, err := cmd.Output()
	if _, ok := err.(*exec.ExitError); ok {
		return "", nil
	}
	if err != nil {
		return fmt.Sprintf("Error running %s %s", name, args), err
	}

	return strings.Trim(string(output), " \n"), nil
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
