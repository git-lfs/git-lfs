package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const sizelimit = 5 * 1024 * 1024

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	oid := latest()
	bad := make(map[string]int64)
	for _, filename := range changed(oid) {
		check(wd, filename, bad)
	}

	if numbad := len(bad); numbad > 0 {
		fmt.Printf("%d bad file(s):\n", numbad)
		for name, size := range bad {
			fmt.Printf("%s %d\n", name, size)
		}
	}
}

func check(working, filename string, bad map[string]int64) {
	full := filepath.Join(working, filename)
	stat, err := os.Lstat(full)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if filesize := stat.Size(); filesize > sizelimit {
		bad[filename] = filesize
	}
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
