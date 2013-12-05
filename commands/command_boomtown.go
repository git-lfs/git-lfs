package gitmedia

import (
	core ".."
	"errors"
)

type BoomtownCommand struct {
	*Command
}

func (c *BoomtownCommand) Run() {
	err := errors.New("Welcome to Boomtown")
	core.Panic(err, "Error!")
}

func init() {
	registerCommand("boomtown", func(c *Command) RunnableCommand {
		return &BoomtownCommand{Command: c}
	})
}
