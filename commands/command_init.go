package gitmedia

import (
	"../gitconfig"
	"fmt"
	"regexp"
)

type InitCommand struct {
	*Command
}

var (
	cleanFilterKey  = "filter.media.clean"
	cleanFilterVal  = "git media clean %f"
	smudgeFilterKey = "filter.media.smudge"
	smudgeFilterVal = "git media smudge %f"
	valueRegexp     = regexp.MustCompile("\\Agit[\\-\\s]media")
)

func (c *InitCommand) Run() {
	clean := gitconfig.Find(cleanFilterKey)
	if shouldReset(clean) {
		fmt.Println("Installing clean filter")
		gitconfig.SetGlobal(cleanFilterKey, cleanFilterVal)
	} else if clean != cleanFilterVal {
		fmt.Printf("Clean filter should be \"%s\" but is \"%s\"\n", cleanFilterVal, clean)
	}

	smudge := gitconfig.Find(smudgeFilterKey)
	if shouldReset(smudge) {
		fmt.Println("Installing smudge filter")
		gitconfig.SetGlobal(smudgeFilterKey, smudgeFilterVal)
	} else if smudge != smudgeFilterVal {
		fmt.Printf("Smudge filter should be \"%s\" but is \"%s\"\n", smudgeFilterVal, smudge)
	}

	fmt.Println("git media initialized")
}

func shouldReset(value string) bool {
	if len(value) == 0 {
		return true
	}
	return valueRegexp.MatchString(value)
}

func init() {
	registerCommand("init", func(c *Command) RunnableCommand {
		return &InitCommand{Command: c}
	})
}
