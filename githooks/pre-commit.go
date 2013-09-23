package main

import (
	"fmt"
	"os"
  "os/exec"
  "strings"
)

func main() {
  oid := latest()
  for _, filename := range changed(oid) {
    fmt.Println(filename)
  }
  os.Exit(1)
}

func changed(oid string) []string {
  lines := simpleExec("git", "diff", "--cached", "--name-only", oid)
  return strings.Split(lines, "\n")
}

func latest() string {
  if oid := simpleExec("git", "rev-parse", "--verify", "HEAD"); oid != "" {
    return oid
  }

  // Initial commit: diff against an empty tree object
  return "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
}

func simpleExec(name string, arg ...string) string {
	output, err := exec.Command(name, arg...).Output()
  if _, ok := err.(*exec.ExitError); ok {
    return ""
  } else if err != nil {
    fmt.Printf("error running: %s %s\n", name, arg)
		panic(err)
	}

	return strings.Trim(string(output), " \n")
}
