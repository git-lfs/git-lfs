package main

import (
	"github.com/github/git-lfs/commands"
	"github.com/github/git-lfs/lfs"
)

func main() {
	commands.Run()
	lfs.LogHttpStats()
}
