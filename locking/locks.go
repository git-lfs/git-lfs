package locking

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/git-lfs/git-lfs/api"
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/tools/kv"
)

var (
	// ErrNoMatchingLocks is an error returned when no matching locks were
	// able to be resolved
	ErrNoMatchingLocks = errors.New("lfs: no matching locks found")
	// ErrLockAmbiguous is an error returned when multiple matching locks
	// were found
	ErrLockAmbiguous = errors.New("lfs: multiple locks found; ambiguous")
)

// Client is the main interface object for the locking package
type Client struct {
	cfg       *config.Configuration
	apiClient *api.Client
	cache     *kv.Store
}

func NewClient(cfg *config.Configuration) (*Client, error) {

	apiClient := api.NewClient(nil)

	lockDir := filepath.Join(config.LocalGitStorageDir, "lfs")
	os.MkdirAll(lockDir, 0755)
	lockFile := filepath.Join(lockDir, "lockcache.db")
	store, err := kv.NewStore(lockFile)
	if err != nil {
		return nil, err
	}
	return &Client{cfg, apiClient, store}, nil
}

// LockFile attempts to lock a file on the given remote name
// path must be relative to the root of the repository
// Returns the lock id if successful, or an error
func (c *Client) LockFile(path, remote string) (id string, e error) {
	// TODO: API currently relies on config.Config but should really pass to client in future
	savedRemote := config.Config.CurrentRemote
	config.Config.CurrentRemote = remote
	defer func() { config.Config.CurrentRemote = savedRemote }()

	// TODO: this is not really the constraint we need to avoid merges, improve as per proposal
	latest, err := git.CurrentRemoteRef()
	if err != nil {
		return "", err
	}

	s, resp := c.apiClient.Locks.Lock(&api.LockRequest{
		Path:               path,
		Committer:          api.CurrentCommitter(),
		LatestRemoteCommit: latest.Sha,
	})

	if _, err := c.apiClient.Do(s); err != nil {
		return "", fmt.Errorf("Error communicating with LFS API: %v", err)
	}

	if len(resp.Err) > 0 {
		return "", fmt.Errorf("Server unable to create lock: %v", resp.Err)
	}

	if err := c.cacheLock(resp.Lock.Path, resp.Lock.Id); err != nil {
		return "", fmt.Errorf("Error caching lock information: %v", err)
	}

	return resp.Lock.Id, nil
}

// UnlockFile attempts to unlock a file on the given remote name
// path must be relative to the root of the repository
// Force causes the file to be unlocked from other users as well
func (c *Client) UnlockFile(path, remote string, force bool) error {
	// TODO: API currently relies on config.Config but should really pass to client in future
	savedRemote := config.Config.CurrentRemote
	config.Config.CurrentRemote = remote
	defer func() { config.Config.CurrentRemote = savedRemote }()

	id, err := c.lockIdFromPath(path)
	if err != nil {
		return fmt.Errorf("Unable to get lock id: %v", err)
	}

	return c.UnlockFileById(id, remote, force)

}

// UnlockFileById attempts to unlock a lock with a given id on the remote
// Force causes the file to be unlocked from other users as well
func (c *Client) UnlockFileById(id, remote string, force bool) error {
	s, resp := c.apiClient.Locks.Unlock(id, force)

	if _, err := c.apiClient.Do(s); err != nil {
		return fmt.Errorf("Error communicating with LFS API: %v", err)
	}

	if len(resp.Err) > 0 {
		return fmt.Errorf("Server unable to unlock lock: %v", resp.Err)
	}

	if err := c.cacheUnlockById(id); err != nil {
		return fmt.Errorf("Error caching unlock information: %v", err)
	}

	return nil
}

// Lock is a record of a locked file
// TODO SJS: review this struct and api equivalent
//           deliberately duplicated to make this API self-contained, could be simpler
type Lock struct {
	// Id is the unique identifier corresponding to this particular Lock. It
	// must be consistent with the local copy, and the server's copy.
	Id string
	// Path is an absolute path to the file that is locked as a part of this
	// lock.
	Path string
	// Name is the name of the person holding this lock
	Name string
	// Email address of the person holding this lock
	Email string
	// CommitSHA is the commit that this Lock was created against. It is
	// strictly equal to the SHA of the minimum commit negotiated in order
	// to create this lock.
	CommitSHA string
	// LockedAt is a required parameter that represents the instant in time
	// that this lock was created. For most server implementations, this
	// should be set to the instant at which the lock was initially
	// received.
	LockedAt time.Time
	// ExpiresAt is an optional parameter that represents the instant in
	// time that the lock stopped being active. If the lock is still active,
	// the server can either a) not send this field, or b) send the
	// zero-value of time.Time.
	UnlockedAt time.Time
}

func (c *Client) newLockFromApi(a api.Lock) Lock {
	return Lock{
		Id:         a.Id,
		Path:       a.Path,
		Name:       a.Committer.Name,
		Email:      a.Committer.Email,
		CommitSHA:  a.CommitSHA,
		LockedAt:   a.LockedAt,
		UnlockedAt: a.UnlockedAt,
	}
}

// SearchLocks returns a channel of locks which match the given name/value filter
// If limit > 0 then search stops at that number of locks
func (c *Client) SearchLocks(remote string, filter map[string]string, limit int) (locks []Lock, err error) {
	// TODO: API currently relies on config.Config but should really pass to client in future
	savedRemote := config.Config.CurrentRemote
	config.Config.CurrentRemote = remote
	defer func() {
		// Only reinstate the remote after we're done
		config.Config.CurrentRemote = savedRemote
	}()

	locks = make([]Lock, 0, limit)

	apifilters := make([]api.Filter, 0, len(filter))
	for k, v := range filter {
		apifilters = append(apifilters, api.Filter{k, v})
	}
	query := &api.LockSearchRequest{Filters: apifilters}
	for {
		s, resp := c.apiClient.Locks.Search(query)
		if _, err := c.apiClient.Do(s); err != nil {
			return locks, fmt.Errorf("Error communicating with LFS API: %v", err)
		}

		if resp.Err != "" {
			return locks, fmt.Errorf("Error response from LFS API: %v", resp.Err)
		}

		for _, l := range resp.Locks {
			locks = append(locks, c.newLockFromApi(l))
			if limit > 0 && len(locks) >= limit {
				// Exit outer loop too
				return locks, nil
			}
		}

		if resp.NextCursor != "" {
			query.Cursor = resp.NextCursor
		} else {
			break
		}
	}

	return locks, nil

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
func (c *Client) lockIdFromPath(path string) (string, error) {
	s, resp := c.apiClient.Locks.Search(&api.LockSearchRequest{
		Filters: []api.Filter{
			{"path", path},
		},
	})

	if _, err := c.apiClient.Do(s); err != nil {
		return "", err
	}

	switch len(resp.Locks) {
	case 0:
		return "", ErrNoMatchingLocks
	case 1:
		return resp.Locks[0].Id, nil
	default:
		return "", ErrLockAmbiguous
	}
}

// Cache a successful lock for faster local lookup later
func (c *Client) cacheLock(filePath, id string) error {
	// TODO
	return nil
}

// Remove a cached lock by path becuase it's been relinquished
func (c *Client) cacheUnlock(filePath string) error {
	// TODO
	return nil
}

// Remove a cached lock by id becuase it's been relinquished
func (c *Client) cacheUnlockById(id string) error {
	// TODO
	return nil
}

// Get the list of cached locked files
func (c *Client) cachedLocks() []string {
	// TODO
	return nil
}

// Fetch locked files for the current committer and cache them locally
// This can be used to sync up locked files when moving machines
func (c *Client) fetchLocksToCache(remoteName string) error {
	// TODO
	return nil
}
