package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/github/git-lfs/commands"
	"github.com/github/git-lfs/lfs"
)

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)

	var once sync.Once

	go func() {
		for {
			sig := <-c
			once.Do(lfs.ClearTempObjects)
			fmt.Fprintf(os.Stderr, "\nExiting because of %q signal.\n", sig)

			exitCode := 1
			if sysSig, ok := sig.(syscall.Signal); ok {
				exitCode = int(sysSig)
			}
			os.Exit(exitCode + 128)
		}
	}()

	commands.Run()
	lfs.LogHttpStats()
	once.Do(lfs.ClearTempObjects)
}
