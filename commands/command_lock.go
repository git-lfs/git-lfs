package commands

import (
	"github.com/github/git-lfs/api"
	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/git"
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

	lockCmd = &cobra.Command{
		Use: "lock",
		Run: lockCommand,
	}
)

func lockCommand(cmd *cobra.Command, args []string) {
	setLockRemoteFor(config.Config)

	if len(args) == 0 {
		Print("Usage: git lfs lock <path>")
		return
	}

	latest, err := git.CurrentRemoteRef()
	if err != nil {
		Error(err.Error())
		Exit("Unable to determine lastest remote ref for branch.")
	}

	s, resp := API.Locks.Lock(&api.LockRequest{
		Path:               args[0],
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

func init() {
	lockCmd.Flags().StringVarP(&lockRemote, "remote", "r", config.Config.CurrentRemote, lockRemoteHelp)

	RootCmd.AddCommand(lockCmd)
}
