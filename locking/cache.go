package locking

import (
	"strings"

	"github.com/git-lfs/git-lfs/tools/kv"
)

const (
	// We want to use a single cache file for integrity, but to make it easy to
	// list all locks, prefix the id->path map in a way we can identify (something
	// that won't be in a path)
	idKeyPrefix string = "*id*://"
)

type LockCache struct {
	kv *kv.Store
}

func NewLockCache(filepath string) (*LockCache, error) {
	kv, err := kv.NewStore(filepath)
	if err != nil {
		return nil, err
	}
	return &LockCache{kv}, nil
}

// Cache a successful lock for faster local lookup later
func (c *LockCache) Add(l Lock) error {
	// Store reference in both directions
	// Path -> Lock
	c.kv.Set(l.Path, &l)
	// EncodedId -> Lock (encoded so we can easily identify)
	c.kv.Set(c.encodeIdKey(l.Id), &l)
	return nil
}

// Remove a cached lock by path becuase it's been relinquished
func (c *LockCache) RemoveByPath(filePath string) error {
	ilock := c.kv.Get(filePath)
	if lock, ok := ilock.(*Lock); ok && lock != nil {
		c.kv.Remove(lock.Path)
		// Id as key is encoded
		c.kv.Remove(c.encodeIdKey(lock.Id))
	}
	return nil
}

// Remove a cached lock by id because it's been relinquished
func (c *LockCache) RemoveById(id string) error {
	// Id as key is encoded
	idkey := c.encodeIdKey(id)
	ilock := c.kv.Get(idkey)
	if lock, ok := ilock.(*Lock); ok && lock != nil {
		c.kv.Remove(idkey)
		c.kv.Remove(lock.Path)
	}
	return nil
}

// Get the list of cached locked files
func (c *LockCache) Locks() []Lock {
	var locks []Lock
	c.kv.Visit(func(key string, val interface{}) bool {
		// Only report file->id entries not reverse
		if !c.isIdKey(key) {
			lock := val.(*Lock)
			locks = append(locks, *lock)
		}
		return true // continue
	})
	return locks
}

// Clear the cache
func (c *LockCache) Clear() {
	c.kv.RemoveAll()
}

// Save the cache
func (c *LockCache) Save() error {
	return c.kv.Save()
}

func (c *LockCache) encodeIdKey(id string) string {
	// Safety against accidents
	if !c.isIdKey(id) {
		return idKeyPrefix + id
	}
	return id
}

func (c *LockCache) decodeIdKey(key string) string {
	// Safety against accidents
	if c.isIdKey(key) {
		return key[len(idKeyPrefix):]
	}
	return key
}

func (c *LockCache) isIdKey(key string) bool {
	return strings.HasPrefix(key, idKeyPrefix)
}
