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
		c.hookInit()
	default:
		c.runInit()
	}

	Print("git media initialized")
}

func (c *InitCommand) runInit() {
	c.globalInit()
	if gitmedia.InRepo() {
		c.hookInit()
	}
}

func (c *InitCommand) globalInit() {
	if err := gitmedia.InstallFilters(); err != nil {
		Error(err.Error())
	}
}

func (c *InitCommand) hookInit() {
	if err := gitmedia.InstallHooks(); err != nil {
		Error(err.Error())
	}
}

func init() {
	registerCommand("init", func(c *Command) RunnableCommand {
		return &InitCommand{Command: c}
	})
}
