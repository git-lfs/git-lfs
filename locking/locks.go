package locking

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/filepathfilter"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tools/kv"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/rubyist/tracerx"
)

var (
	// ErrNoMatchingLocks is an error returned when no matching locks were
	// able to be resolved
	ErrNoMatchingLocks = errors.New(tr.Tr.Get("no matching locks found"))
	// ErrLockAmbiguous is an error returned when multiple matching locks
	// were found
	ErrLockAmbiguous = errors.New(tr.Tr.Get("multiple locks found; ambiguous"))
)

type LockCacher interface {
	Add(l Lock) error
	RemoveByPath(filePath string) error
	RemoveById(id string) error
	Locks() []Lock
	Clear()
	Save() error
}

// Client is the main interface object for the locking package
type Client struct {
	Remote    string
	RemoteRef *git.Ref
	client    lockClient
	cache     LockCacher
	cacheDir  string
	cfg       *config.Configuration

	lockablePatterns []string
	lockableFilter   *filepathfilter.Filter
	lockableMutex    sync.Mutex

	LocalWorkingDir          string
	LocalGitDir              string
	SetLockableFilesReadOnly bool
	ModifyIgnoredFiles       bool
}

// NewClient creates a new locking client with the given configuration
// You must call the returned object's `Close` method when you are finished with
// it
func NewClient(remote string, lfsClient *lfsapi.Client, cfg *config.Configuration) (*Client, error) {
	return &Client{
		Remote:             remote,
		client:             newGenericLockClient(lfsClient),
		cache:              &nilLockCacher{},
		cfg:                cfg,
		ModifyIgnoredFiles: lfsClient.GitEnv().Bool("lfs.lockignoredfiles", false),
		LocalWorkingDir:    cfg.LocalWorkingDir(),
		LocalGitDir:        cfg.LocalGitDir(),
	}, nil
}

func (c *Client) SetupFileCache(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return errors.Wrap(err, tr.Tr.Get("lock cache initialization"))
	}

	lockFile := path
	if stat.IsDir() {
		lockFile = filepath.Join(path, "lockcache.db")
	}

	cache, err := NewLockCache(lockFile)
	if err != nil {
		return errors.Wrap(err, tr.Tr.Get("lock cache initialization"))
	}

	c.cache = cache
	c.cacheDir = filepath.Join(path, "cache")
	return nil
}

// Close this client instance; must be called to dispose of resources
func (c *Client) Close() error {
	return c.cache.Save()
}

// LockFile attempts to lock a file on the current remote
// path must be relative to the root of the repository
// Returns the lock id if successful, or an error
func (c *Client) LockFile(path string) (Lock, error) {
	lockRes, _, err := c.client.Lock(c.Remote, &lockRequest{
		Path: path,
		Ref:  &lockRef{Name: c.RemoteRef.Refspec()},
	})
	if err != nil {
		return Lock{}, errors.Wrap(err, tr.Tr.Get("locking API"))
	}

	if len(lockRes.Message) > 0 {
		if len(lockRes.RequestID) > 0 {
			tracerx.Printf("Server Request ID: %s", lockRes.RequestID)
		}
		return Lock{}, errors.New(tr.Tr.Get("server unable to create lock: %s", lockRes.Message))
	}

	lock := *lockRes.Lock
	if err := c.cache.Add(lock); err != nil {
		return Lock{}, errors.Wrap(err, tr.Tr.Get("lock cache"))
	}

	abs, err := c.getAbsolutePath(path)
	if err != nil {
		return Lock{}, errors.Wrap(err, tr.Tr.Get("make lock path absolute"))
	}

	// If the file exists, ensure that it's writeable on return
	if tools.FileExists(abs) {
		if err := tools.SetFileWriteFlag(abs, true); err != nil {
			return Lock{}, errors.Wrap(err, tr.Tr.Get("set file write flag"))
		}
	}

	return lock, nil
}

// getAbsolutePath takes a repository-relative path and makes it absolute.
//
// For instance, given a repository in /usr/local/src/my-repo and a file called
// dir/foo/bar.txt, getAbsolutePath will return:
//
//	/usr/local/src/my-repo/dir/foo/bar.txt
func (c *Client) getAbsolutePath(p string) (string, error) {
	return filepath.Join(c.LocalWorkingDir, p), nil
}

// UnlockFile attempts to unlock a file on the current remote
// path must be relative to the root of the repository
// Force causes the file to be unlocked from other users as well
func (c *Client) UnlockFile(path string, force bool) error {
	id, err := c.lockIdFromPath(path)
	if err != nil {
		return errors.New(tr.Tr.Get("unable to get lock ID: %v", err))
	}

	return c.UnlockFileById(id, force)
}

// UnlockFileById attempts to unlock a lock with a given id on the current remote
// Force causes the file to be unlocked from other users as well
func (c *Client) UnlockFileById(id string, force bool) error {
	unlockRes, _, err := c.client.Unlock(c.RemoteRef, c.Remote, id, force)
	if err != nil {
		return errors.Wrap(err, tr.Tr.Get("locking API"))
	}

	if len(unlockRes.Message) > 0 {
		if len(unlockRes.RequestID) > 0 {
			tracerx.Printf("Server Request ID: %s", unlockRes.RequestID)
		}
		return errors.New(tr.Tr.Get("server unable to unlock: %s", unlockRes.Message))
	}

	if err := c.cache.RemoveById(id); err != nil {
		return errors.New(tr.Tr.Get("error caching unlock information: %v", err))
	}

	if unlockRes.Lock != nil {
		abs, err := c.getAbsolutePath(unlockRes.Lock.Path)
		if err != nil {
			return errors.Wrap(err, tr.Tr.Get("make lock path absolute"))
		}

		// Make non-writeable if required
		if c.SetLockableFilesReadOnly && c.IsFileLockable(unlockRes.Lock.Path) {
			return tools.SetFileWriteFlag(abs, false)
		}
	}

	return nil
}

// Lock is a record of a locked file
type Lock struct {
	// Id is the unique identifier corresponding to this particular Lock. It
	// must be consistent with the local copy, and the server's copy.
	Id string `json:"id"`
	// Path is an absolute path to the file that is locked as a part of this
	// lock.
	Path string `json:"path"`
	// Owner is the identity of the user that created this lock.
	Owner *User `json:"owner,omitempty"`
	// LockedAt is the time at which this lock was acquired.
	LockedAt time.Time `json:"locked_at"`
}

// SearchLocks returns a channel of locks which match the given name/value filter
// If limit > 0 then search stops at that number of locks
// If localOnly = true, don't query the server & report only own local locks
func (c *Client) SearchLocks(filter map[string]string, limit int, localOnly bool, cached bool) ([]Lock, error) {
	if localOnly {
		return c.searchLocalLocks(filter, limit)
	} else if cached {
		if len(filter) > 0 || limit != 0 {
			return []Lock{}, errors.New(tr.Tr.Get("can't search cached locks when filter or limit is set"))
		}

		locks := []Lock{}
		err := c.readLocksFromCacheFile("remote", func(decoder *json.Decoder) error {
			return decoder.Decode(&locks)
		})
		return locks, err
	} else {
		locks, err := c.searchRemoteLocks(filter, limit)
		if err != nil {
			return locks, err
		}

		if len(filter) == 0 && limit == 0 {
			err = c.writeLocksToCacheFile("remote", func(writer io.Writer) error {
				return c.EncodeLocks(locks, writer)
			})
		}

		return locks, err
	}
}

func (c *Client) SearchLocksVerifiable(limit int, cached bool) (ourLocks, theirLocks []Lock, err error) {
	ourLocks = make([]Lock, 0, limit)
	theirLocks = make([]Lock, 0, limit)

	if cached {
		if limit != 0 {
			return []Lock{}, []Lock{}, errors.New(tr.Tr.Get("can't search cached locks when limit is set"))
		}

		locks := &lockVerifiableList{}
		err := c.readLocksFromCacheFile("verifiable", func(decoder *json.Decoder) error {
			return decoder.Decode(&locks)
		})
		return locks.Ours, locks.Theirs, err
	} else {
		var requestRef *lockRef
		if c.RemoteRef != nil {
			requestRef = &lockRef{Name: c.RemoteRef.Refspec()}
		}

		body := &lockVerifiableRequest{
			Ref:   requestRef,
			Limit: limit,
		}

		c.cache.Clear()

		for {
			list, status, err := c.client.SearchVerifiable(c.Remote, body)
			switch status {
			case http.StatusNotFound, http.StatusNotImplemented:
				return ourLocks, theirLocks, errors.NewNotImplementedError(err)
			case http.StatusForbidden:
				return ourLocks, theirLocks, errors.NewAuthError(err)
			}

			if err != nil {
				return ourLocks, theirLocks, err
			}

			if list.Message != "" {
				if len(list.RequestID) > 0 {
					tracerx.Printf("Server Request ID: %s", list.RequestID)
				}
				return ourLocks, theirLocks, errors.New(tr.Tr.Get("server error searching locks: %s", list.Message))
			}

			for _, l := range list.Ours {
				c.cache.Add(l)
				ourLocks = append(ourLocks, l)
				if limit > 0 && (len(ourLocks)+len(theirLocks)) >= limit {
					return ourLocks, theirLocks, nil
				}
			}

			for _, l := range list.Theirs {
				c.cache.Add(l)
				theirLocks = append(theirLocks, l)
				if limit > 0 && (len(ourLocks)+len(theirLocks)) >= limit {
					return ourLocks, theirLocks, nil
				}
			}

			if list.NextCursor != "" {
				body.Cursor = list.NextCursor
			} else {
				break
			}
		}

		if limit == 0 {
			err = c.writeLocksToCacheFile("verifiable", func(writer io.Writer) error {
				return c.EncodeLocksVerifiable(ourLocks, theirLocks, writer)
			})
		}

		return ourLocks, theirLocks, err
	}
}

func (c *Client) searchLocalLocks(filter map[string]string, limit int) ([]Lock, error) {
	cachedlocks := c.cache.Locks()
	path, filterByPath := filter["path"]
	id, filterById := filter["id"]
	lockCount := 0
	locks := make([]Lock, 0, len(cachedlocks))
	for _, l := range cachedlocks {
		// Manually filter by Path/Id
		if (filterByPath && path != l.Path) ||
			(filterById && id != l.Id) {
			continue
		}
		locks = append(locks, l)
		lockCount++
		if limit > 0 && lockCount >= limit {
			break
		}
	}
	return locks, nil
}

func (c *Client) searchRemoteLocks(filter map[string]string, limit int) ([]Lock, error) {
	locks := make([]Lock, 0, limit)

	apifilters := make([]lockFilter, 0, len(filter))
	for k, v := range filter {
		apifilters = append(apifilters, lockFilter{Property: k, Value: v})
	}

	query := &lockSearchRequest{
		Filters: apifilters,
		Limit:   limit,
		Refspec: c.RemoteRef.Refspec(),
	}

	for {
		list, _, err := c.client.Search(c.Remote, query)
		if err != nil {
			return locks, errors.Wrap(err, tr.Tr.Get("locking"))
		}

		if list.Message != "" {
			if len(list.RequestID) > 0 {
				tracerx.Printf("Server Request ID: %s", list.RequestID)
			}
			return locks, errors.New(tr.Tr.Get("server error searching for locks: %s", list.Message))
		}

		for _, l := range list.Locks {
			locks = append(locks, l)
			if limit > 0 && len(locks) >= limit {
				// Exit outer loop too
				return locks, nil
			}
		}

		if list.NextCursor != "" {
			query.Cursor = list.NextCursor
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
// returned. If no locks matched the given path, an error will be returned.
//
// If the API call is successful, and only one lock matches the given filepath,
// then its ID will be returned, along with a value of "nil" for the error.
func (c *Client) lockIdFromPath(path string) (string, error) {
	list, _, err := c.client.Search(c.Remote, &lockSearchRequest{
		Filters: []lockFilter{
			{Property: "path", Value: path},
		},
		Refspec: c.RemoteRef.Refspec(),
	})

	if err != nil {
		return "", err
	}

	switch len(list.Locks) {
	case 0:
		return "", ErrNoMatchingLocks
	case 1:
		return list.Locks[0].Id, nil
	default:
		return "", ErrLockAmbiguous
	}
}

// IsFileLockedByCurrentCommitter returns whether a file is locked by the
// current user, as cached locally
func (c *Client) IsFileLockedByCurrentCommitter(path string) bool {
	filter := map[string]string{"path": path}
	locks, err := c.searchLocalLocks(filter, 1)
	if err != nil {
		tracerx.Printf("Error searching cached locks: %s\nForcing remote search", err)
		locks, _ = c.searchRemoteLocks(filter, 1)
	}
	return len(locks) > 0
}

func init() {
	kv.RegisterTypeForStorage(&Lock{})
}

func (c *Client) prepareCacheDirectory(kind string) (string, error) {
	cacheDir := filepath.Join(c.cacheDir, "locks")
	if c.RemoteRef != nil {
		cacheDir = filepath.Join(cacheDir, c.RemoteRef.Refspec())
	}

	stat, err := os.Stat(cacheDir)
	if err == nil {
		if !stat.IsDir() {
			return cacheDir, errors.New(tr.Tr.Get("inititalization of cache directory %s failed: already exists, but is no directory", cacheDir))
		}
	} else if os.IsNotExist(err) {
		err = tools.MkdirAll(cacheDir, c.cfg)
		if err != nil {
			return cacheDir, errors.Wrap(err, tr.Tr.Get("initiailization of cache directory %s failed: directory creation failed", cacheDir))
		}
	} else {
		return cacheDir, errors.Wrap(err, tr.Tr.Get("initialization of cache directory %s failed", cacheDir))
	}

	return filepath.Join(cacheDir, kind), nil
}

func (c *Client) readLocksFromCacheFile(kind string, decoder func(*json.Decoder) error) error {
	cacheFile, err := c.prepareCacheDirectory(kind)
	if err != nil {
		return err
	}

	_, err = os.Stat(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New(tr.Tr.Get("no cached locks present"))
		}

		return err
	}

	file, err := os.Open(cacheFile)
	if err != nil {
		return err
	}

	defer file.Close()
	return decoder(json.NewDecoder(file))
}

func (c *Client) EncodeLocks(locks []Lock, writer io.Writer) error {
	return json.NewEncoder(writer).Encode(locks)
}

func (c *Client) EncodeLocksVerifiable(ourLocks, theirLocks []Lock, writer io.Writer) error {
	return json.NewEncoder(writer).Encode(&lockVerifiableList{
		Ours:   ourLocks,
		Theirs: theirLocks,
	})
}

func (c *Client) writeLocksToCacheFile(kind string, writer func(io.Writer) error) error {
	cacheFile, err := c.prepareCacheDirectory(kind)
	if err != nil {
		return err
	}

	file, err := os.Create(cacheFile)
	if err != nil {
		return err
	}

	defer file.Close()
	return writer(file)
}

type nilLockCacher struct{}

func (c *nilLockCacher) Add(l Lock) error {
	return nil
}
func (c *nilLockCacher) RemoveByPath(filePath string) error {
	return nil
}
func (c *nilLockCacher) RemoveById(id string) error {
	return nil
}
func (c *nilLockCacher) Locks() []Lock {
	return nil
}
func (c *nilLockCacher) Clear() {}
func (c *nilLockCacher) Save() error {
	return nil
}
