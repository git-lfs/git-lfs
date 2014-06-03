package commands

import (
	"fmt"
	"github.com/github/git-media/gitconfig"
	"github.com/github/git-media/gitmedia"
	"regexp"
)

type InitCommand struct {
	*Command
}

var valueRegexp = regexp.MustCompile("\\Agit[\\-\\s]media")

func (c *InitCommand) Run() {
	var sub string
	if len(c.SubCommands) > 0 {
		sub = c.SubCommands[0]
	}

	switch sub {
	case "hooks":
		if err := c.hookInit(); err != nil {
			gitmedia.Print("%s", err)
			return
		}
	default:
		c.runInit()
	}

	gitmedia.Print("git media initialized")
}

func (c *InitCommand) runInit() {
	c.globalInit()
	if gitmedia.InRepo() {
		c.hookInit()
	}
}

func (c *InitCommand) globalInit() {
	setFilter("clean")
	setFilter("smudge")
	requireFilters()
}

func (c *InitCommand) hookInit() error {
	return gitmedia.InstallHooks()
}

func setFilter(filterName string) {
	key := fmt.Sprintf("filter.media.%s", filterName)
	value := fmt.Sprintf("git media %s %%f", filterName)

	existing := gitconfig.Find(key)
	if shouldReset(existing) {
		gitmedia.Print("Installing %s filter", filterName)
		gitconfig.UnsetGlobal(key)
		gitconfig.SetGlobal(key, value)
	} else if existing != value {
		gitmedia.Print("The %s filter should be \"%s\" but is \"%s\"", filterName, value, existing)
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
		gitmedia.Print("Media filters should be required but are not")
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
