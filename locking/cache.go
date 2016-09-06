package locking

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/github/git-lfs/api"

	"github.com/boltdb/bolt"
	"github.com/github/git-lfs/config"
)

var (
	dbInit             sync.Once
	lockDb             *bolt.DB
	pathToIdBucketName = []byte("pathToId")
	idToPathBucketName = []byte("idToPath")
)

func initDb(cfg *config.Configuration) error {
	// Open on demand - bolt will lock this file & other processes trying to access it
	// we'll wait a max 5 seconds
	// TODO: could have option to open read-only to take shared lock
	var initerr error
	dbInit.Do(func() {
		lockDir := filepath.Join(config.LocalGitStorageDir, "lfs")
		os.MkdirAll(lockDir, 0755)
		lockFile := filepath.Join(lockDir, "db.lock")
		lockDb, initerr = bolt.Open(lockFile, 0644, &bolt.Options{Timeout: 5 * time.Second})
	})
	if initerr != nil {
		// TODO maybe suggest re-initialising lock cache
		return initerr
	}

	// Very important that Cleanup() is called before shutdown

	return nil
}

func Cleanup() error {
	if lockDb != nil {
		return lockDb.Close()
	}
	return nil
}

// Run a read-only lock database function
// Deals with initialisation and function is a transaction
// Can run in a goroutine if needed
func runLockDbReadOnlyFunc(cfg *config.Configuration, f func(tx *bolt.Tx) error) error {

	if err := initDb(cfg); err != nil {
		return err
	}

	return lockDb.View(f)
}

// Run a read-write lock database function
// Deals with initialisation and function is a transaction
// Can run in a goroutine if needed
func runLockDbFunc(cfg *config.Configuration, f func(tx *bolt.Tx) error) error {

	if err := initDb(cfg); err != nil {
		return err
	}

	// Use Batch() to improve write performance and goroutine friendly
	return lockDb.Batch(f)
}

// This file caches active locks locally so that we can more easily retrieve
// a list of locally locked files without consulting the server
// This only includes locks which the local committer has taken, not all locks

// Cache a successful lock for faster local lookup later
func cacheLock(cfg *config.Configuration, filePath, id string) error {
	return runLockDbFunc(cfg, func(tx *bolt.Tx) error {
		path2id, err := tx.CreateBucketIfNotExists(pathToIdBucketName)
		if err != nil {
			return err
		}
		id2path, err := tx.CreateBucketIfNotExists(idToPathBucketName)
		if err != nil {
			return err
		}
		// Store path -> id and id -> path
		if err := path2id.Put([]byte(filePath), []byte(id)); err != nil {
			return err
		}
		return id2path.Put([]byte(id), []byte(filePath))
	})
}

// Remove a cached lock by path becuase it's been relinquished
func cacheUnlock(cfg *config.Configuration, filePath string) error {
	return runLockDbFunc(cfg, func(tx *bolt.Tx) error {
		path2id := tx.Bucket(pathToIdBucketName)
		id2path := tx.Bucket(idToPathBucketName)
		if path2id == nil || id2path == nil {
			return nil
		}
		idbytes := path2id.Get([]byte(filePath))
		if idbytes != nil {
			if err := id2path.Delete(idbytes); err != nil {
				return err
			}
			return path2id.Delete([]byte(filePath))
		}
		return nil
	})
}

// Remove a cached lock by id becuase it's been relinquished
func cacheUnlockById(cfg *config.Configuration, id string) error {
	return runLockDbFunc(cfg, func(tx *bolt.Tx) error {
		path2id := tx.Bucket(pathToIdBucketName)
		id2path := tx.Bucket(idToPathBucketName)
		if path2id == nil || id2path == nil {
			return nil
		}
		pathbytes := id2path.Get([]byte(id))
		if pathbytes != nil {
			if err := path2id.Delete(pathbytes); err != nil {
				return err
			}
			return id2path.Delete([]byte(id))
		}
		return nil
	})
}

type CachedLock struct {
	Path string
	Id   string
}

// Get the list of cached locked files
func cachedLocks(cfg *config.Configuration) []CachedLock {
	var ret []CachedLock
	runLockDbReadOnlyFunc(cfg, func(tx *bolt.Tx) error {
		path2id := tx.Bucket(pathToIdBucketName)
		if path2id == nil {
			return nil
		}
		path2id.ForEach(func(k []byte, v []byte) error {
			ret = append(ret, CachedLock{string(k), string(v)})
			return nil
		})
		return nil
	})
	return ret
}

// Fetch locked files for the current committer and cache them locally
// This can be used to sync up locked files when moving machines
func fetchLocksToCache(cfg *config.Configuration, remoteName string) error {

	// TODO: filters don't seem to currently define how to search for a
	// committer's email. Is it "committer.email"? For now, just iterate
	lockWrapper := SearchLocks(remoteName, nil, 0)
	var locks []CachedLock
	email := api.CurrentCommitter().Email
	for l := range lockWrapper.Results {
		if l.Committer.Email == email {
			locks = append(locks, CachedLock{l.Path, l.Id})
		}
	}
	err := lockWrapper.Wait()

	if err != nil {
		return err
	}

	// replace cached locks (only do this if search was OK)
	return runLockDbFunc(cfg, func(tx *bolt.Tx) error {
		// Ignore errors deleting buckets
		tx.DeleteBucket(pathToIdBucketName)
		tx.DeleteBucket(idToPathBucketName)
		path2id, err := tx.CreateBucket(pathToIdBucketName)
		if err != nil {
			return err
		}
		id2path, err := tx.CreateBucket(idToPathBucketName)
		if err != nil {
			return err
		}
		for _, l := range locks {
			path2id.Put([]byte(l.Path), []byte(l.Id))
			id2path.Put([]byte(l.Id), []byte(l.Path))
		}
		return nil
	})
}
