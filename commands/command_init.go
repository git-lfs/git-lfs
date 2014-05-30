package gitmedia

import (
	"../gitconfig"
	"fmt"
	"regexp"
)

type InitCommand struct {
	*Command
}

var valueRegexp = regexp.MustCompile("\\Agit[\\-\\s]media")

func (c *InitCommand) Run() {
	setFilter("clean")
	setFilter("smudge")
	requireFilters()

	fmt.Println("git media initialized")
}

func setFilter(filterName string) {
	key := fmt.Sprintf("filter.media.%s", filterName)
	value := fmt.Sprintf("git media %s %%f", filterName)

	existing := gitconfig.Find(key)
	if shouldReset(existing) {
		fmt.Printf("Installing %s filter\n", filterName)
		gitconfig.UnsetGlobal(key)
		gitconfig.SetGlobal(key, value)
	} else if existing != value {
		fmt.Printf("The %s filter should be \"%s\" but is \"%s\"\n", filterName, value, existing)
	}
}

func requireFilters() {
	key := "filter.media.required"
	value := "true"

	existing := gitconfig.Find(key)
	if shouldReset(existing) {
		gitconfig.UnsetGlobal(key)
		gitconfig.SetGlobal(key, value)
	} else if existing != value {
		fmt.Printf("Media filter should be required but are not")
	}
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
