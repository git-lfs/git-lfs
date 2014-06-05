package commands

import (
	"github.com/github/git-media/gitmedia"
)

type VersionCommand struct {
	LovesComics bool
	*Command
}

func (c *VersionCommand) Setup() {
	c.FlagSet.BoolVar(&c.LovesComics, "comics", false, "easter egg")
}

func (c *VersionCommand) Run() {
	Print("%s v%s", c.Name, gitmedia.Version)

	if c.LovesComics {
		Print("Nothing may see Gah Lak Tus and survive.")
	}
}

func init() {
	registerCommand("version", func(c *Command) RunnableCommand {
		return &VersionCommand{Command: c}
	})
}
