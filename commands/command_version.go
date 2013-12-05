package gitmedia

import (
	core ".."
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
		core.Print("Nothing may see Gah Lak Tus and survive.")
	} else {
		core.Print("%s v%s\n", c.Name, core.Version)
	}
}

func init() {
	registerCommand("version", func(c *Command) RunnableCommand {
		return &VersionCommand{Command: c}
	})
}
