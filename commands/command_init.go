package commands

import (
	"fmt"
	"github.com/github/git-media/gitconfig"
	"github.com/github/git-media/gitmedia"
	"io/ioutil"
	"os"
	"path/filepath"
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
	writeHooks()

	fmt.Println("git media initialized")
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
		gitmedia.Print("Media filters should be required but are not.")
	}
}

func shouldReset(value string) bool {
	if len(value) == 0 {
		return true
	}
	return valueRegexp.MatchString(value)
}

func writeHooks() {
	hookPath := filepath.Join(gitmedia.LocalGitDir, "hooks", "pre-push")
	if _, err := os.Stat(hookPath); err == nil {
		gitmedia.Print("Hook already exists: %s", hookPath)
	} else {
		ioutil.WriteFile(hookPath, prePushHook, 0755)
	}
}

var prePushHook = []byte("#!/bin/sh\ngit media push\n")

func init() {
	registerCommand("init", func(c *Command) RunnableCommand {
		return &InitCommand{Command: c}
	})
}
