package gitmedia

import (
	core ".."
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type LogsCommand struct {
	ClearLogs bool
	Boomtown  bool
	*Command
}

func (c *LogsCommand) Setup() {
	c.FlagSet.BoolVar(&c.ClearLogs, "clear", false, "Clear existing error logs")
	c.FlagSet.BoolVar(&c.Boomtown, "boomtown", false, "Trigger a panic")
}

func (c *LogsCommand) Run() {
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
		c.lastLog()
	case "":
		c.listLogs()
	default:
		c.showLog(sub)
	}
}

func (c *LogsCommand) listLogs() {
	for _, path := range sortedLogs() {
		core.Print(path)
	}
}

func (c *LogsCommand) lastLog() {
	logs := sortedLogs()
	c.showLog(logs[len(logs)-1])
}

func (c *LogsCommand) showLog(name string) {
	by, err := ioutil.ReadFile(filepath.Join(core.LocalLogDir, name))
	if err != nil {
		core.Exit("Error reading log: %s", name)
	}

	core.Debug("Reading log: %s", name)
	os.Stdout.Write(by)
}

func (c *LogsCommand) clear() {
	err := os.RemoveAll(core.LocalLogDir)
	if err != nil {
		core.Panic(err, "Error clearing %s", core.LocalLogDir)
	}

	fmt.Println("Cleared", core.LocalLogDir)
}

func (c *LogsCommand) boomtown() {
	core.Debug("Debug message")
	err := errors.New("Error!")
	core.Panic(err, "Welcome to Boomtown")
	core.Debug("Never seen")
}

func sortedLogs() []string {
	fileinfos, err := ioutil.ReadDir(core.LocalLogDir)
	if err != nil {
		return []string{}
	}

	names := make([]string, len(fileinfos))
	for index, info := range fileinfos {
		names[index] = info.Name()
	}

	return names
}

func init() {
	registerCommand("logs", func(c *Command) RunnableCommand {
		return &LogsCommand{Command: c}
	})
}
