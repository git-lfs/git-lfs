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
	core.Print("%s v%s", c.Name, core.Version)

	if c.LovesComics {
		core.Print("Nothing may see Gah Lak Tus and survive.")
	}
}

func init() {
	registerCommand("version", func(c *Command) RunnableCommand {
		return &VersionCommand{Command: c}
	})
}
