//go:build testtools
// +build testtools

package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Need an argument")
		os.Exit(2)
	}
	st, err := os.Stat(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stat %q: %s", os.Args[1], err)
		os.Exit(3)
	}
	mtime := st.ModTime()
	fmt.Printf("%d.%09d", mtime.Unix(), mtime.Nanosecond())
}
