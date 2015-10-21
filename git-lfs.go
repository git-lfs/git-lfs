package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/github/git-lfs/commands"
	"github.com/github/git-lfs/lfs"
)

func main() {
	c := make(chan os.Signal)
	signal.Notify(c)

	var once sync.Once

	go func() {
		for {
			sig := <-c
			once.Do(lfs.ClearTempObjects)
			switch sig {
			case os.Interrupt, os.Kill:
				fmt.Fprintf(os.Stderr, "\nExiting because of %q signal.\n", sig)
				if commands.Erroring {
					os.Exit(1)
				}
				os.Exit(0)
			}
		}
	}()

	commands.Run()
	lfs.LogHttpStats()
	once.Do(lfs.ClearTempObjects)
}
