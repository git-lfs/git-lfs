package main

import (
	".."
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

	if filesize := stat.Size(); int(filesize) > gitmedia.LargeSizeThreshold {
		bad[filename] = filesize
	}
}

func changed(oid string) []string {
	output := gitmedia.SimpleExec("git", "diff-index", "--name-only", oid, "-z")
	files := strings.Split(output, "\x00")
	return files[0 : len(files)-1]
}

func latest() string {
	if oid := gitmedia.SimpleExec("git", "rev-parse", "--verify", "HEAD"); oid != "" {
		return oid
	}

	// Initial commit: diff against an empty tree object
	return "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
}
