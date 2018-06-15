//go:generate goversioninfo

package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/git-lfs/git-lfs/commands"
)

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)

	var once sync.Once

	go func() {
		for {
			sig := <-c
			once.Do(commands.Cleanup)
			fmt.Fprintf(os.Stderr, "\nExiting because of %q signal.\n", sig)

			exitCode := 1
			if sysSig, ok := sig.(syscall.Signal); ok {
				exitCode = int(sysSig)
			}
			os.Exit(exitCode + 128)
		}
	}()

	code := commands.Run()
	once.Do(commands.Cleanup)
	os.Exit(code)
}
