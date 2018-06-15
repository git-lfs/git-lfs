package commands

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/locking"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/spf13/cobra"
)

var (
	locksCmdFlags = new(locksFlags)
)

func locksCommand(cmd *cobra.Command, args []string) {
	filters, err := locksCmdFlags.Filters()
	if err != nil {
		Exit("Error building filters: %v", err)
	}

	if len(lockRemote) > 0 {
		cfg.SetRemote(lockRemote)
	}

	refUpdate := git.NewRefUpdate(cfg.Git, cfg.PushRemote(), cfg.CurrentRef(), nil)
	lockClient := newLockClient()
	lockClient.RemoteRef = refUpdate.Right()
	defer lockClient.Close()

	locks, err := lockClient.SearchLocks(filters, locksCmdFlags.Limit, locksCmdFlags.Local)
	// Print any we got before exiting

	if locksCmdFlags.JSON {
		if err := json.NewEncoder(os.Stdout).Encode(locks); err != nil {
			Error(err.Error())
		}
		return
	}

	var maxPathLen int
	var maxNameLen int
	lockPaths := make([]string, 0, len(locks))
	locksByPath := make(map[string]locking.Lock)
	for _, lock := range locks {
		lockPaths = append(lockPaths, lock.Path)
		locksByPath[lock.Path] = lock
		maxPathLen = tools.MaxInt(maxPathLen, len(lock.Path))
		if lock.Owner != nil {
			maxNameLen = tools.MaxInt(maxNameLen, len(lock.Owner.Name))
		}
	}

	sort.Strings(lockPaths)
	for _, lockPath := range lockPaths {
		var ownerName string
		lock := locksByPath[lockPath]
		if lock.Owner != nil {
			ownerName = lock.Owner.Name
		}

		pathPadding := tools.MaxInt(maxPathLen-len(lock.Path), 0)
		namePadding := tools.MaxInt(maxNameLen-len(ownerName), 0)
		Print("%s%s\t%s%s\tID:%s", lock.Path, strings.Repeat(" ", pathPadding),
			ownerName, strings.Repeat(" ", namePadding),
			lock.Id,
		)
	}

	if err != nil {
		Exit("Error while retrieving locks: %v", errors.Cause(err))
	}
}

// locksFlags wraps up and holds all of the flags that can be given to the
// `git lfs locks` command.
type locksFlags struct {
	// Path is an optional filter parameter to filter against the lock's
	// path
	Path string
	// Id is an optional filter parameter used to filtere against the lock's
	// ID.
	Id string
	// limit is an optional request parameter sent to the server used to
	// limit the
	Limit int
	// local limits the scope of lock reporting to the locally cached record
	// of locks for the current user & doesn't query the server
	Local bool
	// JSON is an optional parameter to output data in json format.
	JSON bool
}

// Filters produces a filter based on locksFlags instance.
func (l *locksFlags) Filters() (map[string]string, error) {
	filters := make(map[string]string)

	if l.Path != "" {
		path, err := lockPath(l.Path)
		if err != nil {
			return nil, err
		}

		filters["path"] = path
	}
	if l.Id != "" {
		filters["id"] = l.Id
	}

	return filters, nil
}

func init() {
	RegisterCommand("locks", locksCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&lockRemote, "remote", "r", "", lockRemoteHelp)
		cmd.Flags().StringVarP(&locksCmdFlags.Path, "path", "p", "", "filter locks results matching a particular path")
		cmd.Flags().StringVarP(&locksCmdFlags.Id, "id", "i", "", "filter locks results matching a particular ID")
		cmd.Flags().IntVarP(&locksCmdFlags.Limit, "limit", "l", 0, "optional limit for number of results to return")
		cmd.Flags().BoolVarP(&locksCmdFlags.Local, "local", "", false, "only list cached local record of own locks")
		cmd.Flags().BoolVarP(&locksCmdFlags.JSON, "json", "", false, "print output in json")
	})
}
