package commands

import (
	"github.com/git-lfs/git-lfs/api"
	"github.com/spf13/cobra"
)

var (
	locksCmdFlags = new(locksFlags)
)

func locksCommand(cmd *cobra.Command, args []string) {
	setLockRemoteFor(cfg)

	filters, err := locksCmdFlags.Filters()
	if err != nil {
		Error(err.Error())
	}

	var locks []api.Lock

	query := &api.LockSearchRequest{Filters: filters}
	for {
		s, resp := API.Locks.Search(query)
		if _, err := API.Do(s); err != nil {
			Error(err.Error())
			Exit("Error communicating with LFS API.")
		}

<<<<<<< HEAD
		if resp.Err != "" {
			Error(resp.Err)
		}

		locks = append(locks, resp.Locks...)

		if locksCmdFlags.Limit > 0 && len(locks) > locksCmdFlags.Limit {
			locks = locks[:locksCmdFlags.Limit]
			break
		}

		if resp.NextCursor != "" {
			query.Cursor = resp.NextCursor
		} else {
			break
		}
=======
	for _, lock := range locks {
		Print("%s\t%s", lock.Path, lock.Owner)
		lockCount++
>>>>>>> f8a50160... Merge branch 'master' into no-dwarf-tables
	}

	Print("\n%d lock(s) matched query:", len(locks))
	for _, lock := range locks {
		Print("%s\t%s <%s>", lock.Path, lock.Committer.Name, lock.Committer.Email)
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
}

// Filters produces a slice of api.Filter instances based on the internal state
// of this locksFlags instance. The return value of this method is capable (and
// recommend to be used with) the api.LockSearchRequest type.
func (l *locksFlags) Filters() ([]api.Filter, error) {
	filters := make([]api.Filter, 0)

	if l.Path != "" {
		path, err := lockPath(l.Path)
		if err != nil {
			return nil, err
		}

		filters = append(filters, api.Filter{"path", path})
	}
	if l.Id != "" {
		filters = append(filters, api.Filter{"id", l.Id})
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
