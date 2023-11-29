package commands

import (
	"encoding/json"
	"os"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/locking"
	"github.com/git-lfs/git-lfs/v3/tr"
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

type unlockResponse struct {
	Id       string `json:"id,omitempty"`
	Path     string `json:"path,omitempty"`
	Unlocked bool   `json:"unlocked"`
	Reason   string `json:"reason,omitempty"`
}

func handleUnlockError(locks []unlockResponse, id string, path string, err error) []unlockResponse {
	Error(err.Error())
	if locksCmdFlags.JSON {
		locks = append(locks, unlockResponse{
			Id:       id,
			Path:     path,
			Unlocked: false,
			Reason:   err.Error(),
		})
	}
	return locks
}

func unlockCommand(cmd *cobra.Command, args []string) {
	hasPath := len(args) > 0
	hasId := len(unlockCmdFlags.Id) > 0
	if hasPath == hasId {
		// If there is both an `--id` AND a `<path>`, or there is
		// neither, print the usage and quit.
		Exit(tr.Tr.Get("Exactly one of --id or a set of paths must be provided"))
	}

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

	locks := make([]unlockResponse, 0, len(args))
	success := true
	if hasPath {
		for _, pathspec := range args {
			path, err := lockPath(lockData, pathspec)
			if err != nil {
				if !unlockCmdFlags.Force {
					locks = handleUnlockError(locks, "", path, errors.New(tr.Tr.Get("Unable to determine path: %v", err.Error())))
					success = false
					continue
				}
				path = pathspec
			}

			if err := unlockAbortIfFileModified(path); err != nil {
				locks = handleUnlockError(locks, "", path, err)
				success = false
				continue
			}

			err = lockClient.UnlockFile(path, unlockCmdFlags.Force)
			if err != nil {
				locks = handleUnlockError(locks, "", path, errors.Cause(err))
				success = false
				continue
			}

			if !locksCmdFlags.JSON {
				Print(tr.Tr.Get("Unlocked %s", path))
				continue
			}
			locks = append(locks, unlockResponse{
				Path:     path,
				Unlocked: true,
			})
		}
	} else if unlockCmdFlags.Id != "" {
		// This call can early-out
		unlockAbortIfFileModifiedById(unlockCmdFlags.Id, lockClient)

		err := lockClient.UnlockFileById(unlockCmdFlags.Id, unlockCmdFlags.Force)
		if err != nil {
			locks = handleUnlockError(locks, unlockCmdFlags.Id, "", errors.New(tr.Tr.Get("Unable to unlock %v: %v", unlockCmdFlags.Id, errors.Cause(err))))
			success = false
		} else if !locksCmdFlags.JSON {
			Print(tr.Tr.Get("Unlocked Lock %s", unlockCmdFlags.Id))
		} else {
			locks = append(locks, unlockResponse{
				Id:       unlockCmdFlags.Id,
				Unlocked: true,
			})
		}
	} else {
		Exit(tr.Tr.Get("Exactly one of --id or a set of paths must be provided"))
	}

	if locksCmdFlags.JSON {
		if err := json.NewEncoder(os.Stdout).Encode(locks); err != nil {
			Error(err.Error())
		}
	}
	if !success {
		lockClient.Close()
		os.Exit(2)
	}
}

func unlockAbortIfFileModified(path string) error {
	modified, err := git.IsFileModified(path)

	if err != nil {
		if unlockCmdFlags.Force {
			// Since git/git@b9a7d55, `git-status(1)` causes an
			// error when asked about files that don't exist,
			// causing `err != nil`, as above.
			//
			// Unlocking a files that does not exist with
			// --force is OK.
			return nil
		}
		return err
	}

	if modified {
		if unlockCmdFlags.Force {
			// Only a warning
			Error(tr.Tr.Get("warning: unlocking with uncommitted changes because --force"))
		} else {
			return errors.New(tr.Tr.Get("Cannot unlock file with uncommitted changes"))
		}

	}
	return nil
}

func unlockAbortIfFileModifiedById(id string, lockClient *locking.Client) error {
	// Get the path so we can check the status
	filter := map[string]string{"id": id}
	// try local cache first
	locks, _ := lockClient.SearchLocks(filter, 0, true, false)
	if len(locks) == 0 {
		// Fall back on calling server
		locks, _ = lockClient.SearchLocks(filter, 0, false, false)
	}

	if len(locks) == 0 {
		// Don't block if we can't determine the path, may be cleaning up old data
		return nil
	}

	return unlockAbortIfFileModified(locks[0].Path)
}

func init() {
	RegisterCommand("unlock", unlockCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&lockRemote, "remote", "r", "", "specify which remote to use when interacting with locks")
		cmd.Flags().StringVarP(&unlockCmdFlags.Id, "id", "i", "", "unlock a lock by its ID")
		cmd.Flags().BoolVarP(&unlockCmdFlags.Force, "force", "f", false, "forcibly break another user's lock(s)")
		cmd.Flags().BoolVarP(&locksCmdFlags.JSON, "json", "", false, "print output in json")
	})
}
