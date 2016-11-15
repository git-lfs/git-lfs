package commands

import (
	"github.com/github/git-lfs/locking"
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

	var lockCount int
	locks := locking.SearchLocks(lockRemote, filters, locksCmdFlags.Limit)

	for lock := range locks.Results {
		Print("%s\t%s <%s>", lock.Path, lock.Committer.Name, lock.Committer.Email)
		lockCount++
	}
	err = locks.Wait()

	if err != nil {
		Exit("Error while retrieving locks: %v", err)
	}

	Print("\n%d lock(s) matched query.", lockCount)
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
	if !isCommandEnabled(cfg, "locks") {
		return
	}

	RegisterCommand("locks", locksCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&lockRemote, "remote", "r", cfg.CurrentRemote, lockRemoteHelp)
		cmd.Flags().StringVarP(&locksCmdFlags.Path, "path", "p", "", "filter locks results matching a particular path")
		cmd.Flags().StringVarP(&locksCmdFlags.Id, "id", "i", "", "filter locks results matching a particular ID")
		cmd.Flags().IntVarP(&locksCmdFlags.Limit, "limit", "l", 0, "optional limit for number of results to return")
	})
}
