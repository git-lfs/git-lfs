package gitmedia

import (
	core ".."
	"fmt"
)

type VersionCommand struct {
	LovesComics bool
	*Command
}

func (c *VersionCommand) Setup() {
	c.FlagSet.BoolVar(&c.LovesComics, "comics", false, "easter egg")
}

func (c *VersionCommand) Run() {
	if c.LovesComics {
		fmt.Println("Nothing may see Gah Lak Tus and survive.")
	} else {
		fmt.Printf("git-media version %s\n", core.Version)
	}
}

func init() {
	registerCommand("version", func(c *Command) RunnableCommand {
		return &VersionCommand{Command: c}
	})
}
