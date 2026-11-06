//go:build testtools
// +build testtools

package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	f, err := os.Create("execute")
	if err == nil {
		f.Close()
	}

	fmt.Printf("execute: %s\n", strings.Join(os.Args, " "))
	fmt.Fprintf(os.Stderr, "execute: %s\n", strings.Join(os.Args, " "))

	os.Exit(1)
}
