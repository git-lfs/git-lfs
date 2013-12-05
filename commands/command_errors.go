package gitmedia

import (
	core ".."
	"errors"
	"fmt"
	"os"
)

type ErrorsCommand struct {
	ClearLogs bool
	Boomtown  bool
	*Command
}

func (c *ErrorsCommand) Setup() {
	c.FlagSet.BoolVar(&c.ClearLogs, "clear", false, "Clear existing error logs")
	c.FlagSet.BoolVar(&c.Boomtown, "boomtown", false, "Trigger a panic")
}

func (c *ErrorsCommand) Run() {
	if c.ClearLogs {
		c.clear()
	}

	if c.Boomtown {
		c.boomtown()
		return
	}
}

func (c *ErrorsCommand) clear() {
	err := os.RemoveAll(core.LocalLogDir)
	if err != nil {
		core.Panic(err, "Error clearing %s", core.LocalLogDir)
	}

	fmt.Println("Cleared", core.LocalLogDir)
}

func (c *ErrorsCommand) boomtown() {
	core.Debug("Debug message")
	err := errors.New("Error!")
	core.Panic(err, "Welcome to Boomtown")
	core.Debug("Never seen")
}

func init() {
	registerCommand("errors", func(c *Command) RunnableCommand {
		return &ErrorsCommand{Command: c}
	})
}
