package commands

import (
	"github.com/github/git-media/gitmedia"
)

type InitCommand struct {
	*Command
}

func (c *InitCommand) Run() {
	var sub string
	if len(c.SubCommands) > 0 {
		sub = c.SubCommands[0]
	}

	switch sub {
	case "hooks":
		if err := c.hookInit(); err != nil {
			gitmedia.Print("%s", err)
			return
		}
	default:
		c.runInit()
	}

	gitmedia.Print("git media initialized")
}

func (c *InitCommand) runInit() {
	c.globalInit()
	if gitmedia.InRepo() {
		c.hookInit()
	}
}

func (c *InitCommand) globalInit() {
	gitmedia.InstallFilters()
}

func (c *InitCommand) hookInit() error {
	return gitmedia.InstallHooks(true)
}

func init() {
	registerCommand("init", func(c *Command) RunnableCommand {
		return &InitCommand{Command: c}
	})
}
