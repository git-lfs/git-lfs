package commands

import (
	"errors"
	"fmt"
	"github.com/github/git-media/gitmedia"
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
		gitmedia.Print(path)
	}
}

func (c *LogsCommand) lastLog() {
	logs := sortedLogs()
	c.showLog(logs[len(logs)-1])
}

func (c *LogsCommand) showLog(name string) {
	by, err := ioutil.ReadFile(filepath.Join(gitmedia.LocalLogDir, name))
	if err != nil {
		gitmedia.Exit("Error reading log: %s", name)
	}

	gitmedia.Debug("Reading log: %s", name)
	os.Stdout.Write(by)
}

func (c *LogsCommand) clear() {
	err := os.RemoveAll(gitmedia.LocalLogDir)
	if err != nil {
		gitmedia.Panic(err, "Error clearing %s", gitmedia.LocalLogDir)
	}

	fmt.Println("Cleared", gitmedia.LocalLogDir)
}

func (c *LogsCommand) boomtown() {
	gitmedia.Debug("Debug message")
	err := errors.New("Error!")
	gitmedia.Panic(err, "Welcome to Boomtown")
	gitmedia.Debug("Never seen")
}

func sortedLogs() []string {
	fileinfos, err := ioutil.ReadDir(gitmedia.LocalLogDir)
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
