package gitmedia

import (
	"errors"
	"fmt"
	"github.com/github/git-media/gitconfig"
	"regexp"
)

var valueRegexp = regexp.MustCompile("\\Agit[\\-\\s]media")

func InstallHooks() error {
	if !InRepo() {
		return errors.New("Not in a repository")
	}

	return nil
}

func InstallFilters() {
	setFilter("clean")
	setFilter("smudge")
	requireFilters()
}

func setFilter(filterName string) {
	key := fmt.Sprintf("filter.media.%s", filterName)
	value := fmt.Sprintf("git media %s %%f", filterName)

	existing := gitconfig.Find(key)
	if shouldReset(existing) {
		Print("Installing %s filter", filterName)
		gitconfig.UnsetGlobal(key)
		gitconfig.SetGlobal(key, value)
	} else if existing != value {
		Print("The %s filter should be \"%s\" but is \"%s\"", filterName, value, existing)
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
		Print("Media filters should be required but are not")
	}
}

func shouldReset(value string) bool {
	if len(value) == 0 {
		return true
	}
	return valueRegexp.MatchString(value)
}
