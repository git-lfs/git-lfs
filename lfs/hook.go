package lfs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/rubyist/tracerx"
)

var (
	// The basic hook which just calls 'git lfs TYPE'
	hookBaseContent = "#!/bin/sh\ncommand -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting '.git/hooks/{{Command}}'.\\n\"; exit 2; }\ngit lfs {{Command}} \"$@\""
	hookOldContent  = "#!/bin/sh\ncommand -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting .git/hooks/{{Command}}.\\n\"; exit 2; }\ngit lfs {{Command}} \"$@\""
)

// A Hook represents a githook as described in http://git-scm.com/docs/githooks.
// Hooks have a type, which is the type of hook that they are, and a body, which
// represents the thing they will execute when invoked by Git.
type Hook struct {
	Type         string
	Contents     string
	Dir          string
	upgradeables []string
	cfg          *config.Configuration
}

type mismatchError struct {
	msg string
}

func (e *mismatchError) Error() string { return e.msg }

func LoadHooks(hookDir string, cfg *config.Configuration) []*Hook {
	return []*Hook{
		NewStandardHook("pre-push", hookDir, []string{
			"#!/bin/sh\ngit lfs push --stdin $*",
			"#!/bin/sh\ngit lfs push --stdin \"$@\"",
			"#!/bin/sh\ngit lfs pre-push \"$@\"",
			"#!/bin/sh\ncommand -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository has been set up with Git LFS but Git LFS is not installed.\\n\"; exit 0; }\ngit lfs pre-push \"$@\"",
			"#!/bin/sh\ncommand -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository has been set up with Git LFS but Git LFS is not installed.\\n\"; exit 2; }\ngit lfs pre-push \"$@\"",
			hookOldContent,
		}, cfg),
		NewStandardHook("post-checkout", hookDir, []string{hookOldContent}, cfg),
		NewStandardHook("post-commit", hookDir, []string{hookOldContent}, cfg),
		NewStandardHook("post-merge", hookDir, []string{hookOldContent}, cfg),
	}
}

// NewStandardHook creates a new hook using the template script calling 'git lfs theType'
func NewStandardHook(theType, hookDir string, upgradeables []string, cfg *config.Configuration) *Hook {
	formattedUpgradeables := make([]string, 0, len(upgradeables))
	for _, s := range upgradeables {
		formattedUpgradeables = append(formattedUpgradeables, strings.Replace(s, "{{Command}}", theType, -1))
	}
	return &Hook{
		Type:         theType,
		Contents:     strings.Replace(hookBaseContent, "{{Command}}", theType, -1),
		Dir:          hookDir,
		upgradeables: formattedUpgradeables,
		cfg:          cfg,
	}
}

func (h *Hook) Exists() bool {
	_, err := os.Stat(h.Path())

	return !os.IsNotExist(err)
}

// Path returns the desired (or actual, if installed) location where this hook
// should be installed. It returns an absolute path in all cases.
func (h *Hook) Path() string {
	return filepath.Join(h.Dir, h.Type)
}

// Install installs this Git hook on disk, or upgrades it if it does exist, and
// is upgradeable. It will create a hooks directory relative to the local Git
// directory. It returns and halts at any errors, and returns nil if the
// operation was a success.
func (h *Hook) Install(force bool, append bool) error {
	msg := fmt.Sprintf("Install hook: %s, force=%t, path=%s", h.Type, force, h.Path())

	if err := tools.MkdirAll(h.Dir, h.cfg); err != nil {
		return err
	}

	if h.Exists() && !force {
		tracerx.Printf(msg + ", upgrading...")
		return h.Upgrade(append)
	}

	tracerx.Printf(msg)
	return h.write(false)
}

// write writes the contents of this Hook to disk, appending a newline at the
// end, and sets the mode to octal 0755. It writes to disk unconditionally, and
// returns at any error.
func (h *Hook) write(append_ bool) error {
	contents := h.Contents + "\n"
	if append_ {
		s, err := ioutil.ReadFile(h.Path())
		if err != nil {
			return err
		}

		contents = string(s) + strings.TrimPrefix(contents, "#!/bin/sh")
	}

	return ioutil.WriteFile(h.Path(), []byte(contents), 0755)
}

// Upgrade upgrades the (assumed to be) existing git hook to the current
// contents. A hook is considered "upgrade-able" if its contents are matched in
// the member variable `Upgradeables`. It halts and returns any errors as they
// arise.
func (h *Hook) Upgrade(append bool) error {
	match, err := h.matchesCurrent()
	if err != nil {
		if _, ok := err.(*mismatchError); ok && append {
			return h.write(true)
		}

		return err
	}

	if !match {
		return nil
	}

	return h.write(false)
}

// Uninstall removes the hook on disk so long as it matches the current version,
// or any of the past versions of this hook.
func (h *Hook) Uninstall() error {
	msg := fmt.Sprintf("Uninstall hook: %s, path=%s", h.Type, h.Path())

	match, err := h.matchesCurrent()
	if err != nil {
		return err
	}

	if !match {
		tracerx.Printf(msg + ", doesn't match...")
		return nil
	}

	tracerx.Printf(msg)
	return os.RemoveAll(h.Path())
}

// matchesCurrent returns whether or not an existing git hook is able to be
// written to or upgraded. A git hook matches those conditions if and only if
// its contents match the current contents, or any past "upgrade-able" contents
// of this hook.
func (h *Hook) matchesCurrent() (bool, error) {
	file, err := os.Open(h.Path())
	if err != nil {
		return false, err
	}

	by, err := ioutil.ReadAll(io.LimitReader(file, 1024))
	file.Close()
	if err != nil {
		return false, err
	}

	contents := strings.TrimSpace(tools.Undent(string(by)))
	if contents == h.Contents || len(contents) == 0 {
		return true, nil
	}

	for _, u := range h.upgradeables {
		if u == contents {
			return true, nil
		}
	}

	return false, &mismatchError{fmt.Sprintf("%s\n\n%s\n", tr.Tr.Get("Hook already exists: %s", string(h.Type)), tools.Indent(contents))}
}
