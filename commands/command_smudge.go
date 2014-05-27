package gitmedia

import (
	".."
	"../filters"
	"os"
)

type SmudgeCommand struct {
	*Command
}

func (c *SmudgeCommand) Run() {
	sha, err := gitmedia.Decode(os.Stdin)
	if err != nil {
		gitmedia.Panic(err, "Error reading git-media meta data from stdin:")
	}

	err = gitmediafilters.Smudge(os.Stdout, sha)
	if err != nil {
		smudgerr := err.(*gitmediafilters.SmudgeError)
		gitmedia.Panic(err, "Error reading file from local media dir: %s", smudgerr.Filename)
	}
}

func init() {
	registerCommand("smudge", func(c *Command) RunnableCommand {
		return &SmudgeCommand{Command: c}
	})
}
