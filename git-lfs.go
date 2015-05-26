package main

import (
	"github.com/github/git-lfs/commands"

	// Keep this so godep thinks the olekukonko/ts package is necessary. It's only
	// referenced in cheggaaa/pb/pb_win.go
	_ "github.com/github/git-lfs/Godeps/_workspace/src/github.com/olekukonko/ts"
)

func main() {
	commands.Run()
}
