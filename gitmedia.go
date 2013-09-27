package gitmedia

import (
	"fmt"
	"os/exec"
	"strings"
)

const Version = "0.0.1"

var LargeSizeThreshold = 5 * 1024 * 1024

func SimpleExec(name string, arg ...string) string {
	output, err := exec.Command(name, arg...).Output()
	if _, ok := err.(*exec.ExitError); ok {
		return ""
	} else if err != nil {
		fmt.Printf("error running: %s %s\n", name, arg)
		panic(err)
	}

	return strings.Trim(string(output), " \n")
}
