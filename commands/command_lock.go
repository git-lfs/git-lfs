package commands

import (
	"github.com/github/git-lfs/api"
	"github.com/github/git-lfs/git"
	"github.com/spf13/cobra"
)

var (
	lockCmd = &cobra.Command{
		Use: "lock",
		Run: lockCommand,
	}
)

func lockCommand(cmd *cobra.Command, args []string) {
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
	RootCmd.AddCommand(lockCmd)
}
