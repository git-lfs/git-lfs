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
			once.Do(clearTempObjects)
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
	once.Do(clearTempObjects)
}

func clearTempObjects() {
	s := lfs.LocalStorage
	if s == nil {
		return
	}

	if err := s.ClearTempObjects(); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening %q to clear old temp files: %s\n", s.TempDir, err)
	}
}
