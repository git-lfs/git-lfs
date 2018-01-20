package commands

import (
	"encoding/json"
	"os"

	"github.com/git-lfs/git-lfs/errors"
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

var unlockUsage = "Usage: git lfs unlock (--id my-lock-id | <path>)"

func unlockCommand(cmd *cobra.Command, args []string) {
	hasPath := len(args) > 0
	hasId := len(unlockCmdFlags.Id) > 0
	if hasPath == hasId {
		// If there is both an `--id` AND a `<path>`, or there is
		// neither, print the usage and quit.
		Exit(unlockUsage)
	}

	if len(lockRemote) > 0 {
		cfg.SetRemote(lockRemote)
	}

	refUpdate := git.NewRefUpdate(cfg.Git, cfg.PushRemote(), cfg.CurrentRef(), nil)
	lockClient := newLockClient()
	lockClient.RemoteRef = refUpdate.Right()
	defer lockClient.Close()

	if hasPath {
		path, err := lockPath(args[0])
		if err != nil {
			if !unlockCmdFlags.Force {
				Exit("Unable to determine path: %v", err.Error())
			}
			path = args[0]
		}

		// This call can early-out
		unlockAbortIfFileModified(path, !os.IsNotExist(err))

		err = lockClient.UnlockFile(path, unlockCmdFlags.Force)
		if err != nil {
			Exit("%s", errors.Cause(err))
		}

		if !locksCmdFlags.JSON {
			Print("Unlocked %s", path)
			return
		}
	} else if unlockCmdFlags.Id != "" {
		// This call can early-out
		unlockAbortIfFileModifiedById(unlockCmdFlags.Id, lockClient)

		err := lockClient.UnlockFileById(unlockCmdFlags.Id, unlockCmdFlags.Force)
		if err != nil {
			Exit("Unable to unlock %v: %v", unlockCmdFlags.Id, errors.Cause(err))
		}

		if !locksCmdFlags.JSON {
			Print("Unlocked Lock %s", unlockCmdFlags.Id)
			return
		}
	} else {
		Error(unlockUsage)
	}

	if err := json.NewEncoder(os.Stdout).Encode(struct {
		Unlocked bool `json:"unlocked"`
	}{true}); err != nil {
		Error(err.Error())
	}
	return
}

func unlockAbortIfFileModified(path string, exists bool) {
	modified, err := git.IsFileModified(path)

	if err != nil {
		if !exists && unlockCmdFlags.Force {
			// Since git/git@b9a7d55, `git-status(1)` causes an
			// error when asked about files that don't exist,
			// causing `err != nil`, as above.
			//
			// Unlocking a files that does not exist with
			// --force is OK.
			return
		}
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

	unlockAbortIfFileModified(locks[0].Path, true)
}

func init() {
	RegisterCommand("unlock", unlockCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&lockRemote, "remote", "r", "", lockRemoteHelp)
		cmd.Flags().StringVarP(&unlockCmdFlags.Id, "id", "i", "", "unlock a lock by its ID")
		cmd.Flags().BoolVarP(&unlockCmdFlags.Force, "force", "f", false, "forcibly break another user's lock(s)")
		cmd.Flags().BoolVarP(&locksCmdFlags.JSON, "json", "", false, "print output in json")
	})
}
