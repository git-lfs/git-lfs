package gitmedia

type CleanCommand struct {
	*Command
}

type SmudgeCommand struct {
	*Command
}

func (c *CleanCommand) Run() {
	err := PipeMediaCommand("git-media-clean")
	if err != nil {
		panic(err)
	}
}

func (c *SmudgeCommand) Run() {
	err := PipeMediaCommand("git-media-smudge")
	if err != nil {
		panic(err)
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
