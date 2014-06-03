package commands

import (
	"github.com/github/git-media/filters"
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/metafile"
	"os"
)

type SmudgeCommand struct {
	*Command
}

func (c *SmudgeCommand) Run() {
	sha, err := metafile.Decode(os.Stdin)
	if err != nil {
		gitmedia.Panic(err, "Error reading git-media meta data from stdin:")
	}

	err = filters.Smudge(os.Stdout, sha)
	if err != nil {
		smudgerr := err.(*filters.SmudgeError)
		gitmedia.Panic(err, "Error reading file from local media dir: %s", smudgerr.Filename)
	}
}

func init() {
	registerCommand("smudge", func(c *Command) RunnableCommand {
		return &SmudgeCommand{Command: c}
	})
}
