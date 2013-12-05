package gitmedia

import (
	core ".."
	"errors"
)

type BoomtownCommand struct {
	*Command
}

func (c *BoomtownCommand) Run() {
	core.Debug("Debug message")
	err := errors.New("Error!")
	core.Panic(err, "Welcome to Boomtown")
	core.Debug("Never seen")
}

func init() {
	registerCommand("boomtown", func(c *Command) RunnableCommand {
		return &BoomtownCommand{Command: c}
	})
}
