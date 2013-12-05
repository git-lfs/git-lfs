package gitmedia

import (
	".."
)

type CleanCommand struct {
	*Command
}

type SmudgeCommand struct {
	*Command
}

func (c *CleanCommand) Run() {
	err := PipeMediaCommand("git-media-clean")
	if err != nil {
		gitmedia.Panic(err, "Error running 'git media clean'")
	}
}

func (c *SmudgeCommand) Run() {
	err := PipeMediaCommand("git-media-smudge")
	if err != nil {
		gitmedia.Panic(err, "Error running 'git media smudge'")
	}
}

func init() {
	registerCommand("clean", func(c *Command) RunnableCommand {
		return &CleanCommand{Command: c}
	})

	registerCommand("smudge", func(c *Command) RunnableCommand {
		return &SmudgeCommand{Command: c}
	})
}
