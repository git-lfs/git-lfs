package git

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func simpleExec(stdin io.Reader, name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	if stdin != nil {
		cmd.Stdin = stdin
	}

	output, err := cmd.Output()
	if _, ok := err.(*exec.ExitError); ok {
		return "", nil
	} else if err != nil {
		return fmt.Sprintf("Error running %s %s", name, arg), err
	}

	return strings.Trim(string(output), " \n"), nil
}
