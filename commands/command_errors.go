package gitmedia

import (
	core ".."
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

	var sub string
	if len(c.SubCommands) > 0 {
		sub = c.SubCommands[0]
	}

	switch sub {
	case "last":
		c.lastError()
	case "":
		c.listErrors()
	default:
		core.Exit("Invalid errors sub command: %s", sub)
	}
}

func (c *ErrorsCommand) listErrors() {
	for _, path := range sortedLogs() {
		core.Print(path)
	}
}

func (c *ErrorsCommand) lastError() {
	logs := sortedLogs()
	last := logs[len(logs)-1]
	by, err := ioutil.ReadFile(filepath.Join(core.LocalLogDir, last))
	if err != nil {
		core.Panic(err, "Error reading log: %s", last)
	}

	core.Debug("Reading log: %s", last)
	os.Stdout.Write(by)
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

func sortedLogs() []string {
	fileinfos, err := ioutil.ReadDir(core.LocalLogDir)
	if err != nil {
		core.Panic(err, "Error reading logs directory: %s", core.LocalLogDir)
	}

	names := make([]string, len(fileinfos))
	for index, info := range fileinfos {
		names[index] = info.Name()
	}

	return names
}

func init() {
	registerCommand("errors", func(c *Command) RunnableCommand {
		return &ErrorsCommand{Command: c}
	})
}
