package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/locking"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/spf13/cobra"
)

var (
	lockRemote string
)

func lockCommand(cmd *cobra.Command, args []string) {
	if len(lockRemote) > 0 {
		cfg.SetRemote(lockRemote)
	}

	lockData, err := computeLockData()
	if err != nil {
		ExitWithError(err)
	}

	refUpdate := git.NewRefUpdate(cfg.Git, cfg.PushRemote(), cfg.CurrentRef(), nil)
	lockClient := newLockClient()
	lockClient.RemoteRef = refUpdate.RemoteRef()
	defer lockClient.Close()

	success := true
	locks := make([]locking.Lock, 0, len(args))
	for _, path := range args {
		path, err := lockPath(lockData, path)
		if err != nil {
			Error(err.Error())
			success = false
			continue
		}

		lock, err := lockClient.LockFile(path)
		if err != nil {
			Error(tr.Tr.Get("Locking %s failed: %v", path, errors.Cause(err)))
			success = false
			continue
		}

		locks = append(locks, lock)

		if locksCmdFlags.JSON {
			continue
		}

		Print(tr.Tr.Get("Locked %s", path))
	}

	if locksCmdFlags.JSON {
		if err := json.NewEncoder(os.Stdout).Encode(locks); err != nil {
			Error(err.Error())
			success = false
		}
	}

	if !success {
		lockClient.Close()
		os.Exit(2)
	}
}

type lockData struct {
	rootDir    string
	workingDir string
}

// computeLockData computes data about the given repository and working
// directory to use in lockPath.
func computeLockData() (*lockData, error) {
	wd, err := tools.Getwd()
	if err != nil {
		return nil, err
	}
	wd, err = tools.CanonicalizeSystemPath(wd)
	if err != nil {
		return nil, err
	}
	return &lockData{
		rootDir:    cfg.LocalWorkingDir(),
		workingDir: wd,
	}, nil
}

// lockPaths relativizes the given filepath such that it is relative to the root
// path of the repository it is contained within, taking into account the
// working directory of the caller.
//
// lockPaths also respects different filesystem directory separators, so that a
// Windows path of "\foo\bar" will be normalized to "foo/bar".
//
// If the file path cannot be determined, an error will be returned. If the file
// in question is actually a directory, an error will be returned. Otherwise,
// the cleaned path will be returned.
//
// For example:
//   - Working directory: /code/foo/bar/
//   - Repository root: /code/foo/
//   - File to lock: ./baz
//   - Resolved path bar/baz
func lockPath(data *lockData, file string) (string, error) {
	var abs string
	var err error

	if filepath.IsAbs(file) {
		abs, err = tools.CanonicalizeSystemPath(file)
		if err != nil {
			return "", errors.New(tr.Tr.Get("unable to canonicalize path %q: %v", file, err))
		}
	} else {
		abs = filepath.Join(data.workingDir, file)
	}
	path, err := filepath.Rel(data.rootDir, abs)
	if err != nil {
		return "", err
	}

	path = filepath.ToSlash(path)
	if strings.HasPrefix(path, "../") {
		return "", errors.New(tr.Tr.Get("unable to canonicalize path %q", path))
	}

	if stat, err := os.Stat(abs); err == nil && stat.IsDir() {
		return path, errors.New(tr.Tr.Get("cannot lock directory: %s", file))
	}

	return filepath.ToSlash(path), nil
}

func init() {
	RegisterCommand("lock", lockCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&lockRemote, "remote", "r", "", "specify which remote to use when interacting with locks")
		cmd.Flags().BoolVarP(&locksCmdFlags.JSON, "json", "", false, "print output in json")
	})
}
