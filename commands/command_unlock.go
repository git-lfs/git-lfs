package commands

import (
	"errors"

<<<<<<< HEAD
	"github.com/git-lfs/git-lfs/api"
=======
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/locking"
>>>>>>> f8a50160... Merge branch 'master' into no-dwarf-tables
	"github.com/spf13/cobra"
)

var (
	// errNoMatchingLocks is an error returned when no matching locks were
	// able to be resolved
	errNoMatchingLocks = errors.New("lfs: no matching locks found")
	// errLockAmbiguous is an error returned when multiple matching locks
	// were found
	errLockAmbiguous = errors.New("lfs: multiple locks found; ambiguous")

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
	setLockRemoteFor(cfg)

	var id string
	if len(args) != 0 {
		path, err := lockPath(args[0])
<<<<<<< HEAD
		if err != nil {
			Error(err.Error())
		}

		if id, err = lockIdFromPath(path); err != nil {
			Error(err.Error())
		}
	} else if unlockCmdFlags.Id != "" {
		id = unlockCmdFlags.Id
=======
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
>>>>>>> f8a50160... Merge branch 'master' into no-dwarf-tables
	} else {
		Error("Usage: git lfs unlock (--id my-lock-id | <path>)")
	}

	s, resp := API.Locks.Unlock(id, unlockCmdFlags.Force)

	if _, err := API.Do(s); err != nil {
		Error(err.Error())
		Exit("Error communicating with LFS API.")
	}

	if len(resp.Err) > 0 {
		Error(resp.Err)
		Exit("Server unable to unlock lock.")
	}

	Print("'%s' was unlocked (%s)", args[0], resp.Lock.Id)
}

// lockIdFromPath makes a call to the LFS API and resolves the ID for the locked
// locked at the given path.
//
// If the API call failed, an error will be returned. If multiple locks matched
// the given path (should not happen during real-world usage), an error will be
// returnd. If no locks matched the given path, an error will be returned.
//
// If the API call is successful, and only one lock matches the given filepath,
// then its ID will be returned, along with a value of "nil" for the error.
func lockIdFromPath(path string) (string, error) {
	s, resp := API.Locks.Search(&api.LockSearchRequest{
		Filters: []api.Filter{
			{"path", path},
		},
	})

	if _, err := API.Do(s); err != nil {
		return "", err
	}

	switch len(resp.Locks) {
	case 0:
		return "", errNoMatchingLocks
	case 1:
		return resp.Locks[0].Id, nil
	default:
		return "", errLockAmbiguous
	}
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
	if !isCommandEnabled(cfg, "locks") {
		return
	}

	RegisterCommand("unlock", unlockCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&lockRemote, "remote", "r", cfg.CurrentRemote, lockRemoteHelp)
		cmd.Flags().StringVarP(&unlockCmdFlags.Id, "id", "i", "", "unlock a lock by its ID")
		cmd.Flags().BoolVarP(&unlockCmdFlags.Force, "force", "f", false, "forcibly break another user's lock(s)")
	})
}
