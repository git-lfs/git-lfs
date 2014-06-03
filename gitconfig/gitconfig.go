package gitconfig

import (
	"fmt"
	"os/exec"
	"strings"
)

func Find(val string) string {
	output, _ := SimpleExec("git", "config", val)
	return output
}

func SetGlobal(key, val string) {
	SimpleExec("git", "config", "--global", "--add", key, val)
}

func UnsetGlobal(key string) {
	SimpleExec("git", "config", "--global", "--unset", key)
}

func List() (string, error) {
	return SimpleExec("git", "config", "-l")
}

func ListFromFile() (string, error) {
	return SimpleExec("git", "config", "-l", "-f", ".gitconfig")
}

func SimpleExec(name string, arg ...string) (string, error) {
	output, err := exec.Command(name, arg...).Output()
	if _, ok := err.(*exec.ExitError); ok {
		return "", nil
	} else if err != nil {
		return fmt.Sprintf("Error running %s %s", name, arg), err
	}

	return strings.Trim(string(output), " \n"), nil
}
