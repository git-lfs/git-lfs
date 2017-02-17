package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/api"
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/tools/longpathos"
	"github.com/spf13/cobra"
)

var (
	lockRemote     string
	lockRemoteHelp = "specify which remote to use when interacting with locks"

	// TODO(taylor): consider making this (and the above flag) a property of
	// some parent-command, or another similarly less ugly way of handling
	// this
	setLockRemoteFor = func(c *config.Configuration) {
		c.CurrentRemote = lockRemote
	}
)

func lockCommand(cmd *cobra.Command, args []string) {
	setLockRemoteFor(cfg)

	if len(args) == 0 {
		Print("Usage: git lfs lock <path>")
		return
	}

	latest, err := git.CurrentRemoteRef()
	if err != nil {
		Error(err.Error())
		Exit("Unable to determine lastest remote ref for branch.")
	}

	path, err := lockPath(args[0])
	if err != nil {
		Exit(err.Error())
	}

	s, resp := API.Locks.Lock(&api.LockRequest{
		Path:               path,
		Committer:          api.CurrentCommitter(),
		LatestRemoteCommit: latest.Sha,
	})

	if _, err := API.Do(s); err != nil {
		Error(err.Error())
		Exit("Error communicating with LFS API.")
	}

	if len(resp.Err) > 0 {
		Error(resp.Err)
		Exit("Server unable to create lock.")
	}

	Print("\n'%s' was locked (%s)", args[0], resp.Lock.Id)
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
<<<<<<< HEAD

	if stat, err := longpathos.Stat(abs); err != nil {
		return "", err
=======
	path = strings.TrimPrefix(path, string(os.PathSeparator))
	if stat, err := os.Stat(abs); err != nil {
		return path, err
>>>>>>> f8a50160... Merge branch 'master' into no-dwarf-tables
	} else {
		if stat.IsDir() {
			return path, fmt.Errorf("lfs: cannot lock directory: %s", file)
		}

		return path, nil
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
