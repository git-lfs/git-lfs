package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
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

	if len(lockRemote) > 0 {
		cfg.SetRemote(lockRemote)
	}

	refUpdate := git.NewRefUpdate(cfg.Git, cfg.PushRemote(), cfg.CurrentRef(), nil)
	lockClient := newLockClient()
	lockClient.RemoteRef = refUpdate.Right()
	defer lockClient.Close()

	lock, err := lockClient.LockFile(path)
	if err != nil {
		Exit("Lock failed: %v", errors.Cause(err))
	}

	if locksCmdFlags.JSON {
		if err := json.NewEncoder(os.Stdout).Encode(lock); err != nil {
			Error(err.Error())
		}
		return
	}

	Print("Locked %s", path)
}

// lockPaths relativizes the given filepath such that it is relative to the root
// path of the repository it is contained within, taking into account the
// working directory of the caller.
//
// lockPaths also respects different filesystem directory separators, so that a
// Windows path of "\foo\bar" will be normalized to "foo/bar".
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
	wd, err = filepath.EvalSymlinks(wd)
	if err != nil {
		return "", errors.Wrapf(err,
			"could not follow symlinks for %s", wd)
	}

	abs := filepath.Join(wd, file)
	path := strings.TrimPrefix(abs, repo)
	path = strings.TrimPrefix(path, string(os.PathSeparator))
	if stat, err := os.Stat(abs); err != nil {
		return "", err
	} else {
		if stat.IsDir() {
			return path, fmt.Errorf("lfs: cannot lock directory: %s", file)
		}

		return filepath.ToSlash(path), nil
	}
}

func init() {
	RegisterCommand("lock", lockCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&lockRemote, "remote", "r", "", lockRemoteHelp)
		cmd.Flags().BoolVarP(&locksCmdFlags.JSON, "json", "", false, "print output in json")
	})
}
