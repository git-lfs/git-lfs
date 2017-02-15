package commands

import (
	"encoding/json"
	"os"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/locking"
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
	lockClient := newLockClient(lockRemote)
	defer lockClient.Close()

	if len(args) != 0 {
		path, err := lockPath(args[0])
		if err != nil && !unlockCmdFlags.Force {
			Exit("Unable to determine path: %v", err.Error())
		}

		// This call can early-out
		unlockAbortIfFileModified(path)

		err = lockClient.UnlockFile(path, unlockCmdFlags.Force)
		if err != nil {
			Exit("Unable to unlock: %v", err.Error())
		}
	} else if unlockCmdFlags.Id != "" {

		// This call can early-out
		unlockAbortIfFileModifiedById(unlockCmdFlags.Id, lockClient)

		err := lockClient.UnlockFileById(unlockCmdFlags.Id, unlockCmdFlags.Force)
		if err != nil {
			Exit("Unable to unlock %v: %v", unlockCmdFlags.Id, err.Error())
		}
	} else {
		Error("Usage: git lfs unlock (--id my-lock-id | <path>)")
	}

	if locksCmdFlags.JSON {
		if err := json.NewEncoder(os.Stdout).Encode(struct {
			Unlocked bool `json:"unlocked"`
		}{true}); err != nil {
			Error(err.Error())
		}
		return
	}
	Print("'%s' was unlocked", args[0])
}

func unlockAbortIfFileModified(path string) {
	modified, err := git.IsFileModified(path)

	if err != nil {
		Exit(err.Error())
	}

	if modified {
		if unlockCmdFlags.Force {
			// Only a warning
			Error("Warning: unlocking with uncommitted changes because --force")
		} else {
			Exit("Cannot unlock file with uncommitted changes")
		}

	}
}

func unlockAbortIfFileModifiedById(id string, lockClient *locking.Client) {
	// Get the path so we can check the status
	filter := map[string]string{"id": id}
	// try local cache first
	locks, _ := lockClient.SearchLocks(filter, 0, true)
	if len(locks) == 0 {
		// Fall back on calling server
		locks, _ = lockClient.SearchLocks(filter, 0, false)
	}

	if len(locks) == 0 {
		// Don't block if we can't determine the path, may be cleaning up old data
		return
	}

	unlockAbortIfFileModified(locks[0].Path)

}

func init() {
	RegisterCommand("unlock", unlockCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&lockRemote, "remote", "r", cfg.CurrentRemote, lockRemoteHelp)
		cmd.Flags().StringVarP(&unlockCmdFlags.Id, "id", "i", "", "unlock a lock by its ID")
		cmd.Flags().BoolVarP(&unlockCmdFlags.Force, "force", "f", false, "forcibly break another user's lock(s)")
		cmd.Flags().BoolVarP(&locksCmdFlags.JSON, "json", "", false, "print output in json")
	})
}
