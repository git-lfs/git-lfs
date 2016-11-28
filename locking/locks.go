package locking

import (
	"errors"
	"fmt"

	"github.com/git-lfs/git-lfs/api"
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/tools"
)

var (
	// API is a package-local instance of the API client for use within
	// various command implementations.
	apiClient = api.NewClient(nil)
	// errNoMatchingLocks is an error returned when no matching locks were
	// able to be resolved
	errNoMatchingLocks = errors.New("lfs: no matching locks found")
	// errLockAmbiguous is an error returned when multiple matching locks
	// were found
	errLockAmbiguous = errors.New("lfs: multiple locks found; ambiguous")
)

// Lock attempts to lock a file on the given remote name
// path must be relative to the root of the repository
// Returns the lock id if successful, or an error
func Lock(path, remote string) (id string, e error) {
	// TODO: API currently relies on config.Config but should really pass to client in future
	savedRemote := config.Config.CurrentRemote
	config.Config.CurrentRemote = remote
	defer func() { config.Config.CurrentRemote = savedRemote }()

	// TODO: this is not really the constraint we need to avoid merges, improve as per proposal
	latest, err := git.CurrentRemoteRef()
	if err != nil {
		return "", err
	}

	s, resp := apiClient.Locks.Lock(&api.LockRequest{
		Path:               path,
		Committer:          api.CurrentCommitter(),
		LatestRemoteCommit: latest.Sha,
	})

	if _, err := apiClient.Do(s); err != nil {
		return "", fmt.Errorf("Error communicating with LFS API: %v", err)
	}

	if len(resp.Err) > 0 {
		return "", fmt.Errorf("Server unable to create lock: %v", resp.Err)
	}

	if err := cacheLock(resp.Lock.Path, resp.Lock.Id); err != nil {
		return "", fmt.Errorf("Error caching lock information: %v", err)
	}

	return resp.Lock.Id, nil
}

// Unlock attempts to unlock a file on the given remote name
// path must be relative to the root of the repository
// Force causes the file to be unlocked from other users as well
func Unlock(path, remote string, force bool) error {
	// TODO: API currently relies on config.Config but should really pass to client in future
	savedRemote := config.Config.CurrentRemote
	config.Config.CurrentRemote = remote
	defer func() { config.Config.CurrentRemote = savedRemote }()

	id, err := lockIdFromPath(path)
	if err != nil {
		return fmt.Errorf("Unable to get lock id: %v", err)
	}

	return UnlockById(id, remote, force)

}

// Unlock attempts to unlock a lock with a given id on the remote
// Force causes the file to be unlocked from other users as well
func UnlockById(id, remote string, force bool) error {
	s, resp := apiClient.Locks.Unlock(id, force)

	if _, err := apiClient.Do(s); err != nil {
		return fmt.Errorf("Error communicating with LFS API: %v", err)
	}

	if len(resp.Err) > 0 {
		return fmt.Errorf("Server unable to unlock lock: %v", resp.Err)
	}

	if err := cacheUnlockById(id); err != nil {
		return fmt.Errorf("Error caching unlock information: %v", err)
	}

	return nil
}

// ChannelWrapper for lock search to more easily return async error data via Wait()
// See NewPointerChannelWrapper for construction / use
type LockChannelWrapper struct {
	*tools.BaseChannelWrapper
	Results <-chan api.Lock
}

// Construct a new channel wrapper for api.Lock
// Caller can use s.Results directly for normal processing then call Wait() to finish & check for errors
func NewLockChannelWrapper(lockChan <-chan api.Lock, errChan <-chan error) *LockChannelWrapper {
	return &LockChannelWrapper{tools.NewBaseChannelWrapper(errChan), lockChan}
}

// SearchLocks returns a channel of locks which match the given name/value filter
// If limit > 0 then search stops at that number of locks
func SearchLocks(remote string, filter map[string]string, limit int) *LockChannelWrapper {
	// TODO: API currently relies on config.Config but should really pass to client in future
	savedRemote := config.Config.CurrentRemote
	config.Config.CurrentRemote = remote
	errChan := make(chan error, 5) // can be multiple errors below
	lockChan := make(chan api.Lock, 10)
	c := NewLockChannelWrapper(lockChan, errChan)
	go func() {
		defer func() {
			close(lockChan)
			close(errChan)
			// Only reinstate the remote after we're done
			config.Config.CurrentRemote = savedRemote
		}()

		apifilters := make([]api.Filter, 0, len(filter))
		for k, v := range filter {
			apifilters = append(apifilters, api.Filter{k, v})
		}
		lockCount := 0
		query := &api.LockSearchRequest{Filters: apifilters}
	QueryLoop:
		for {
			s, resp := apiClient.Locks.Search(query)
			if _, err := apiClient.Do(s); err != nil {
				errChan <- fmt.Errorf("Error communicating with LFS API: %v", err)
				break
			}

			if resp.Err != "" {
				errChan <- fmt.Errorf("Error response from LFS API: %v", resp.Err)
				break
			}

			for _, l := range resp.Locks {
				lockChan <- l
				lockCount++
				if limit > 0 && lockCount >= limit {
					// Exit outer loop too
					break QueryLoop
				}
			}

			if resp.NextCursor != "" {
				query.Cursor = resp.NextCursor
			} else {
				break
			}
		}

	}()

	return c

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
	s, resp := apiClient.Locks.Search(&api.LockSearchRequest{
		Filters: []api.Filter{
			{"path", path},
		},
	})

	if _, err := apiClient.Do(s); err != nil {
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
