package gitmedia

import (
	"errors"
	"fmt"
	"github.com/github/git-media/gitconfig"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

var (
	valueRegexp           = regexp.MustCompile("\\Agit[\\-\\s]media")
	prePushHook           = []byte("#!/bin/sh\ngit media push\n")
	NotInARepositoryError = errors.New("Not in a repository")
)

type HookExists struct {
	Name string
	Path string
}

func (e *HookExists) Error() string {
	return fmt.Sprintf("Hook already exists: %s", e.Name)
}

func InstallHooks() error {
	if !InRepo() {
		return NotInARepositoryError
	}

	if err := os.MkdirAll(filepath.Join(LocalGitDir, "hooks"), 0755); err != nil {
		return err
	}

	hookPath := filepath.Join(LocalGitDir, "hooks", "pre-push")
	if _, err := os.Stat(hookPath); err == nil {
		return &HookExists{"pre-push", hookPath}
	} else {
		return ioutil.WriteFile(hookPath, prePushHook, 0755)
	}

	return nil
}

func InstallFilters() error {
	var err error
	err = setFilter("clean")
	if err == nil {
		err = setFilter("smudge")
	}
	if err == nil {
		err = requireFilters()
	}
	return err
}

func setFilter(filterName string) error {
	key := fmt.Sprintf("filter.media.%s", filterName)
	value := fmt.Sprintf("git media %s %%f", filterName)

	existing := gitconfig.Find(key)
	if shouldReset(existing) {
		gitconfig.UnsetGlobal(key)
		gitconfig.SetGlobal(key, value)
	} else if existing != value {
		return fmt.Errorf("The %s filter should be \"%s\" but is \"%s\"", filterName, value, existing)
	}

	return nil
}

func requireFilters() error {
	key := "filter.media.required"
	value := "true"

	existing := gitconfig.Find(key)
	if shouldReset(existing) {
		gitconfig.UnsetGlobal(key)
		gitconfig.SetGlobal(key, value)
	} else if existing != value {
		return errors.New("Media filters should be required but are not.")
	}

	return nil
}

func shouldReset(value string) bool {
	if len(value) == 0 {
		return true
	}
	return valueRegexp.MatchString(value)
}
