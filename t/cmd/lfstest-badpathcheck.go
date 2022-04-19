//go:build testtools
// +build testtools

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("exploit")
	fmt.Fprintln(os.Stderr, "exploit")

	f, err := os.Create("exploit")
	if err != nil {
		f.Close()
	}
}
