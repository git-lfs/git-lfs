package main

import (
	"github.com/github/git-lfs/commands"
	"github.com/github/git-lfs/lfs"
)

//go:generate go run docs/include-help-text.go

func main() {
	commands.Run()
	lfs.LogHttpStats()
}
