package locking

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/github/git-lfs/config"
	"github.com/stretchr/testify/assert"
)

func testLockCache(t *testing.T) {
	var err error

	oldStore := config.LocalGitStorageDir
	config.LocalGitStorageDir, err = ioutil.TempDir("", "testCacheLock")
	defer func() {
		Cleanup()
		os.RemoveAll(config.LocalGitStorageDir)
		config.LocalGitStorageDir = oldStore
	}()
	assert.Nil(t, err)

	cfg := &config.Configuration{}
	err = cacheLock(cfg, "folder/test1.dat", "101")
	assert.Nil(t, err)
	err = cacheLock(cfg, "folder/test2.dat", "102")
	assert.Nil(t, err)
	err = cacheLock(cfg, "root.dat", "103")
	assert.Nil(t, err)

	locks := cachedLocks(cfg)
	assert.Equal(t, []CachedLock{
		CachedLock{"folder/test1.dat", "101"},
		CachedLock{"folder/test2.dat", "102"},
		CachedLock{"root.dat", "103"},
	}, locks)

	err = cacheUnlock(cfg, "folder/test2.dat")
	assert.Nil(t, err)

	locks = cachedLocks(cfg)
	assert.Equal(t, []CachedLock{
		CachedLock{"folder/test1.dat", "101"},
		CachedLock{"root.dat", "103"},
	}, locks)

	err = cacheUnlockById(cfg, "101")
	assert.Nil(t, err)

	locks = cachedLocks(cfg)
	assert.Equal(t, []CachedLock{
		CachedLock{"root.dat", "103"},
	}, locks)
}
