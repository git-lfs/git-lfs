package lfs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// A Hook represents a githook as described in http://git-scm.com/docs/githooks.
// Hooks have a type, which is the type of hook that they are, and a body, which
// represents the thing they will execute when invoked by Git.
type Hook struct {
	Type         string
	Contents     string
	Upgradeables []string
}

func (h *Hook) Exists() bool {
	_, err := os.Stat(h.Path())
	return err == nil
}

// Path returns the desired (or actual, if installed) location where this hook
// should be installed, relative to the local Git directory.
func (h *Hook) Path() string {
	return filepath.Join(LocalGitDir, "hooks", string(h.Type))
}

// Install installs this Git hook on disk, or upgrades it if it does exist, and
// is upgradeable. It will create a hooks directory relative to the local Git
// directory. It returns and halts at any errors, and returns nil if the
// operation was a success.
func (h *Hook) Install(force bool) error {
	if err := os.MkdirAll(filepath.Join(LocalGitDir, "hooks"), 0755); err != nil {
		return err
	}

	if h.Exists() && !force {
		return h.Upgrade()
	}

	return h.write()
}

// write writes the contents of this Hook to disk, appending a newline at the
// end, and sets the mode to octal 0755. It writes to disk unconditionally, and
// returns at any error.
func (h *Hook) write() error {
	return ioutil.WriteFile(h.Path(), []byte(h.Contents+"\n"), 0755)
}

// Upgrade upgrades the (assumed to be) existing git hook to the current
// contents. A hook is considered "upgrade-able" if its contents are matched in
// the member variable `Upgradeables`. It halts and returns any errors as they
// arise.
func (h *Hook) Upgrade() error {
	match, err := h.matchesCurrent()
	if err != nil {
		return err
	}

	if !match {
		return nil
	}

	return h.write()
}

// Uninstall removes the hook on disk so long as it matches the current version,
// or any of the past versions of this hook.
func (h *Hook) Uninstall() error {
	if !InRepo() {
		return newInvalidRepoError(nil)
	}

	match, err := h.matchesCurrent()
	if err != nil {
		return err
	}

	if !match {
		return nil
	}

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

	contents := strings.TrimSpace(string(by))
	if contents == h.Contents {
		return true, nil
	}

	for _, u := range h.Upgradeables {
		if u == contents {
			return true, nil
		}
	}

	return false, fmt.Errorf("Hook already exists: %s\n\n%s\n", string(h.Type), contents)
}
