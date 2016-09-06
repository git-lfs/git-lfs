package locking

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/github/git-lfs/api"
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

type TestLifecycle struct {
}

func (l *TestLifecycle) Build(schema *api.RequestSchema) (*http.Request, error) {
	return http.NewRequest("GET", "http://dummy", nil)
}

func (l *TestLifecycle) Execute(req *http.Request, into interface{}) (api.Response, error) {
	// Return test data including other users
	locks := api.LockList{Locks: []api.Lock{
		api.Lock{Id: "99", Path: "folder/test3.dat", Committer: api.Committer{Name: "Alice", Email: "alice@wonderland.com"}},
		api.Lock{Id: "101", Path: "folder/test1.dat", Committer: api.Committer{Name: "Fred", Email: "fred@bloggs.com"}},
		api.Lock{Id: "102", Path: "folder/test2.dat", Committer: api.Committer{Name: "Fred", Email: "fred@bloggs.com"}},
		api.Lock{Id: "103", Path: "root.dat", Committer: api.Committer{Name: "Fred", Email: "fred@bloggs.com"}},
		api.Lock{Id: "199", Path: "other/test1.dat", Committer: api.Committer{Name: "Charles", Email: "charles@incharge.com"}},
	}}
	locksJson, _ := json.Marshal(locks)
	r := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.0",
		Body:       ioutil.NopCloser(bytes.NewReader(locksJson)),
	}
	if into != nil {
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(into); err != nil {
			return nil, err
		}
	}
	return api.WrapHttpResponse(r), nil
}
func (l *TestLifecycle) Cleanup(resp api.Response) error {
	return resp.Body().Close()
}

func testRefreshCache(t *testing.T) {
	var err error
	oldStore := config.LocalGitStorageDir
	config.LocalGitStorageDir, err = ioutil.TempDir("", "testCacheLock")
	defer func() {
		Cleanup()
		os.RemoveAll(config.LocalGitStorageDir)
		config.LocalGitStorageDir = oldStore
	}()
	assert.Nil(t, err)

	// api.CurrentCommitter reads from global config so have to change it
	oldConfig := config.Config
	config.Config = config.NewFrom(config.Values{
		Git: map[string]string{"user.name": "Fred", "user.email": "fred@bloggs.com"}})
	defer func() {
		config.Config = oldConfig
	}()
	oldClient := API
	API = api.NewClient(&TestLifecycle{})
	defer func() {
		API = oldClient
	}()

	// Should start with no cached items
	cfg := &config.Configuration{}
	locks := cachedLocks(cfg)
	assert.Empty(t, locks)

	// Should load from test data, just Fred's
	err = fetchLocksToCache(cfg, "origin")
	assert.Nil(t, err)

	locks = cachedLocks(cfg)
	assert.Equal(t, []CachedLock{
		CachedLock{"folder/test1.dat", "101"},
		CachedLock{"folder/test2.dat", "102"},
		CachedLock{"root.dat", "103"},
	}, locks)

}
