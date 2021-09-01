package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/locking"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/spf13/cobra"
)

var (
	lockRemote     string
	lockRemoteHelp = "specify which remote to use when interacting with locks"
)

func lockCommand(cmd *cobra.Command, args []string) {
	if len(lockRemote) > 0 {
		cfg.SetRemote(lockRemote)
	}

	refUpdate := git.NewRefUpdate(cfg.Git, cfg.PushRemote(), cfg.CurrentRef(), nil)
	lockClient := newLockClient()
	lockClient.RemoteRef = refUpdate.Right()
	defer lockClient.Close()

	success := true
	locks := make([]locking.Lock, 0, len(args))
	for _, path := range args {
		path, err := lockPath(path)
		if err != nil {
			Error(err.Error())
			success = false
			continue
		}

		lock, err := lockClient.LockFile(path)
		if err != nil {
			Error("Locking %s failed: %v", path, errors.Cause(err))
			success = false
			continue
		}

		locks = append(locks, lock)

		if locksCmdFlags.JSON {
			continue
		}

		Print("Locked %s", path)
	}

	if locksCmdFlags.JSON {
		if err := json.NewEncoder(os.Stdout).Encode(locks); err != nil {
			Error(err.Error())
			success = false
		}
	}

	if !success {
		os.Exit(2)
	}
}

// lockPaths relativizes the given filepath such that it is relative to the root
// path of the repository it is contained within, taking into account the
// working directory of the caller.
//
// lockPaths also respects different filesystem directory separators, so that a
// Windows path of "\foo\bar" will be normalized to "foo/bar".
//
// If the root directory, working directory, or file cannot be
// determined, an error will be returned. If the file in question is
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
	wd, err = tools.CanonicalizeSystemPath(wd)
	if err != nil {
		return "", errors.Wrapf(err,
			"could not follow symlinks for %s", wd)
	}

	var abs string
	if filepath.IsAbs(file) {
		abs, err = tools.CanonicalizeSystemPath(file)
		if err != nil {
			return "", fmt.Errorf("lfs: unable to canonicalize path %q: %v", file, err)
		}
	} else {
		abs = filepath.Join(wd, file)
	}
	path, err := filepath.Rel(repo, abs)
	if err != nil {
		return "", err
	}

	path = filepath.ToSlash(path)
	if strings.HasPrefix(path, "../") {
		return "", fmt.Errorf("lfs: unable to canonicalize path %q", path)
	}

	if stat, err := os.Stat(abs); err == nil && stat.IsDir() {
		return path, fmt.Errorf("lfs: cannot lock directory: %s", file)
	}

	return filepath.ToSlash(path), nil
}

func init() {
	RegisterCommand("lock", lockCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&lockRemote, "remote", "r", "", lockRemoteHelp)
		cmd.Flags().BoolVarP(&locksCmdFlags.JSON, "json", "", false, "print output in json")
	})
}
