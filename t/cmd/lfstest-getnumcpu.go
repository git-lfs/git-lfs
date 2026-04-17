//go:build testtools
// +build testtools

package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Print(runtime.NumCPU())
}
