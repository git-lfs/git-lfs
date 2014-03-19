package gitmedia

import (
	"../gitconfig"
	"fmt"
)

type InitCommand struct {
	*Command
}

var (
	cleanFilterKey  = "filter.media.clean"
	cleanFilterVal  = "git-media-clean %f"
	smudgeFilterKey = "filter.media.smudge"
	smudgeFilterVal = "git-media-smudge %f"
)

func (c *InitCommand) Run() {
	clean := gitconfig.Find(cleanFilterKey)
	if clean == "" {
		fmt.Println("Installing clean filter")
		gitconfig.SetGlobal(cleanFilterKey, cleanFilterVal)
	} else if clean != cleanFilterVal {
		fmt.Printf("Clean filter should be \"%s\" but is \"%s\"\n", cleanFilterVal, clean)
	}

	smudge := gitconfig.Find(smudgeFilterKey)
	if smudge == "" {
		fmt.Println("Installing smudge filter")
		gitconfig.SetGlobal(smudgeFilterKey, smudgeFilterVal)
	} else if smudge != smudgeFilterVal {
		fmt.Printf("Smudge filter should be \"%s\" but is \"%s\"\n", smudgeFilterVal, smudge)
	}

	fmt.Println("git media initialized")
}

func init() {
	registerCommand("init", func(c *Command) RunnableCommand {
		return &InitCommand{Command: c}
	})
}
