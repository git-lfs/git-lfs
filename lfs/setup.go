package lfs

import (
	"errors"
	"fmt"
	"github.com/github/git-lfs/git"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	valueRegexp           = regexp.MustCompile("\\Agit[\\-\\s]media")
	NotInARepositoryError = errors.New("Not in a repository")

	prePushHook     = "#!/bin/sh\ngit lfs pre-push \"$@\""
	prePushUpgrades = map[string]bool{
		"#!/bin/sh\ngit lfs push --stdin $*":     true,
		"#!/bin/sh\ngit lfs push --stdin \"$@\"": true,
	}
)

type HookExists struct {
	Name     string
	Path     string
	Contents string
}

func (e *HookExists) Error() string {
	return fmt.Sprintf("Hook already exists: %s\n\n%s\n", e.Name, e.Contents)
}

func InstallHooks(force bool) error {
	if !InRepo() {
		return NotInARepositoryError
	}

	if err := os.MkdirAll(filepath.Join(LocalGitDir, "hooks"), 0755); err != nil {
		return err
	}

	hookPath := filepath.Join(LocalGitDir, "hooks", "pre-push")
	if _, err := os.Stat(hookPath); err == nil && !force {
		return upgradeHookOrError(hookPath, "pre-push", prePushHook, prePushUpgrades)
	}

	return ioutil.WriteFile(hookPath, []byte(prePushHook+"\n"), 0755)
}

func upgradeHookOrError(hookPath, hookName, hook string, upgrades map[string]bool) error {
	file, err := os.Open(hookPath)
	if err != nil {
		return err
	}

	by, err := ioutil.ReadAll(io.LimitReader(file, 1024))
	file.Close()
	if err != nil {
		return err
	}

	contents := strings.TrimSpace(string(by))
	if contents == hook {
		return nil
	}

	if upgrades[contents] {
		return ioutil.WriteFile(hookPath, []byte(hook+"\n"), 0755)
	}

	return &HookExists{hookName, hookPath, contents}
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
	key := fmt.Sprintf("filter.lfs.%s", filterName)
	value := fmt.Sprintf("git lfs %s %%f", filterName)

	existing := git.Config.Find(key)
	if shouldReset(existing) {
		git.Config.UnsetGlobal(key)
		git.Config.SetGlobal(key, value)
	} else if existing != value {
		return fmt.Errorf("The %s filter should be \"%s\" but is \"%s\"", filterName, value, existing)
	}

	return nil
}

func requireFilters() error {
	key := "filter.lfs.required"
	value := "true"

	existing := git.Config.Find(key)
	if shouldReset(existing) {
		git.Config.UnsetGlobal(key)
		git.Config.SetGlobal(key, value)
	} else if existing != value {
		return errors.New("Git LFS filters should be required but are not.")
	}

	return nil
}

func shouldReset(value string) bool {
	if len(value) == 0 {
		return true
	}
	return valueRegexp.MatchString(value)
}
