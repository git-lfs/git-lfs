package gitconfig

import (
	"fmt"
	"os/exec"
	"strings"
)

func Find(val string) string {
	output, _ := simpleExec("git", "config", val)
	return output
}

func SetGlobal(key, val string) {
	simpleExec("git", "config", "--global", "--add", key, val)
}

func UnsetGlobal(key string) {
	simpleExec("git", "config", "--global", "--unset", key)
}

func List() (string, error) {
	return simpleExec("git", "config", "-l")
}

func ListFromFile() (string, error) {
	return simpleExec("git", "config", "-l", "-f", ".gitconfig")
}

func simpleExec(name string, arg ...string) (string, error) {
	output, err := exec.Command(name, arg...).Output()
	if _, ok := err.(*exec.ExitError); ok {
		return "", nil
	} else if err != nil {
		return fmt.Sprintf("Error running %s %s", name, arg), err
	}

	return strings.Trim(string(output), " \n"), nil
}
