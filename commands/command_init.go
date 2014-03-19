package gitmedia

import (
	"../gitconfig"
	"fmt"
)

type InitCommand struct {
	*Command
}

func (c *InitCommand) Run() {
	verifyAndSet("filter.media.clean", "git-media-clean %f", "clean filter")
	verifyAndSet("filter.media.smudge", "git-media-smudge %f", "smudge filter")

	fmt.Println("git media initialized")
}

func init() {
	registerCommand("init", func(c *Command) RunnableCommand {
		return &InitCommand{Command: c}
	})
}

func verifyAndSet(key, val, name string) {
	current := gitconfig.Find(key)
	if current == "" {
		fmt.Println("Setting up", name)
		gitconfig.SetGlobal(key, val)
	}
}
