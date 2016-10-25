package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/locking"
	"github.com/spf13/cobra"
)

var (
	lockRemote     string
	lockRemoteHelp = "specify which remote to use when interacting with locks"
)

func lockCommand(cmd *cobra.Command, args []string) {

	if len(args) == 0 {
		Print("Usage: git lfs lock <path>")
		return
	}

	path, err := lockPath(args[0])
	if err != nil {
		Exit(err.Error())
	}

	id, err := locking.Lock(path, lockRemote)
	if err != nil {
		Exit("Lock failed: %v", err)
	}

	Print("\n'%s' was locked (%s)", args[0], id)
}

// lockPaths relativizes the given filepath such that it is relative to the root
// path of the repository it is contained within, taking into account the
// working directory of the caller.
//
// If the root directory, working directory, or file cannot be
// determined/opened, an error will be returned. If the file in question is
// actually a directory, an error will be returned. Otherwise, the cleaned path
// will be returned.
//
// For example:
//     - Working directory: /code/foo/bar/
//     - Repository root: /code/foo/
//     - File to lock: ./baz
//     - Resolved path bar/baz
func lockPath(file string) (string, error) {
	repo, err := git.RootDir()
	if err != nil {
		return "", err
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	abs := filepath.Join(wd, file)
	path := strings.TrimPrefix(abs, repo)

	if stat, err := os.Stat(abs); err != nil {
		return "", err
	} else {
		if stat.IsDir() {
			return "", fmt.Errorf("lfs: cannot lock directory: %s", file)
		}

		return path[1:], nil
	}
}

func init() {
	if !isCommandEnabled(cfg, "locks") {
		return
	}

	RegisterCommand("lock", lockCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&lockRemote, "remote", "r", cfg.CurrentRemote, lockRemoteHelp)
	})
}
