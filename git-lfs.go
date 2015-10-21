package main

import (
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
			<-c // We don't care which os.Signal was received.
			once.Do(lfs.ClearTempObjects)
		}
	}()

	commands.Run()
	lfs.LogHttpStats()
	once.Do(lfs.ClearTempObjects)
}
