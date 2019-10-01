// +build testtools

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func canonicalize(path string) (string, error) {
	left := path
	right := ""

	for {
		canon, err := filepath.EvalSymlinks(left)
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		if err == nil {
			if right == "" {
				return canon, nil
			}
			return filepath.Join(canon, right), nil
		}
		// One component of our path is missing.  Let's walk up a level
		// and canonicalize that and then append the remaining piece.
		full := filepath.Join(left, right)
		if right == "" {
			full = left
		}
		newleft := filepath.Clean(fmt.Sprintf("%s%c..", left, os.PathSeparator))
		newright, err := filepath.Rel(newleft, full)
		if err != nil {
			return "", err
		}
		left = newleft
		right = newright
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s PATH\n", os.Args[0])
		os.Exit(2)
	}

	path, err := filepath.Abs(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating absolute path: %v", err)
		os.Exit(3)
	}

	fullpath, err := canonicalize(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error canonicalizing: %v", err)
		os.Exit(4)
	}

	fmt.Println(fullpath)
}
