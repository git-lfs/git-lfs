package commands

import (
	"github.com/github/git-lfs/locking"

	"github.com/spf13/cobra"
)

var (
	unlockCmdFlags unlockFlags
)

// unlockFlags holds the flags given to the `git lfs unlock` command
type unlockFlags struct {
	// Id is the Id of the lock that is being unlocked.
	Id string
	// Force specifies whether or not the `lfs unlock` command was invoked
	// with "--force", signifying the user's intent to break another
	// individual's lock(s).
	Force bool
}

func unlockCommand(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		path, err := lockPath(args[0])
		if err != nil {
			Exit("Unable to determine path: %v", err.Error())
		}

		err = locking.Unlock(path, lockRemote, unlockCmdFlags.Force)
		if err != nil {
			Exit("Unable to unlock: %v", err.Error())
		}
	} else if unlockCmdFlags.Id != "" {
		err := locking.UnlockById(unlockCmdFlags.Id, lockRemote, unlockCmdFlags.Force)
		if err != nil {
			Exit("Unable to unlock %v: %v", unlockCmdFlags.Id, err.Error())
		}
	} else {
		Error("Usage: git lfs unlock (--id my-lock-id | <path>)")
	}

	Print("'%s' was unlocked", args[0])
}

func init() {
	if !isCommandEnabled(cfg, "locks") {
		return
	}

	RegisterCommand("unlock", unlockCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&lockRemote, "remote", "r", cfg.CurrentRemote, lockRemoteHelp)
		cmd.Flags().StringVarP(&unlockCmdFlags.Id, "id", "i", "", "unlock a lock by its ID")
		cmd.Flags().BoolVarP(&unlockCmdFlags.Force, "force", "f", false, "forcibly break another user's lock(s)")
	})
}
