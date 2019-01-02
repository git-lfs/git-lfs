package locking

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/git-lfs/git-lfs/lfshttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type LocksById []Lock

func (a LocksById) Len() int           { return len(a) }
func (a LocksById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a LocksById) Less(i, j int) bool { return a[i].Id < a[j].Id }

func TestRemoteLocksWithCache(t *testing.T) {
	var err error
	tempDir, err := ioutil.TempDir("", "testCacheLock")
	assert.Nil(t, err)

	remoteQueries := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteQueries++

		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/locks", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(&lockList{
			Locks: []Lock{
				Lock{Id: "100", Path: "folder/test1.dat", Owner: &User{Name: "Alice"}},
				Lock{Id: "101", Path: "folder/test2.dat", Owner: &User{Name: "Charles"}},
				Lock{Id: "102", Path: "folder/test3.dat", Owner: &User{Name: "Fred"}},
			},
		})
		assert.Nil(t, err)
	}))

	defer func() {
		srv.Close()
	}()

	lfsclient, err := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.url":    srv.URL + "/api",
		"user.name":  "Fred",
		"user.email": "fred@bloggs.com",
	}))
	require.Nil(t, err)

	client, err := NewClient("", lfsclient, config.New())
	assert.Nil(t, err)
	assert.Nil(t, client.SetupFileCache(tempDir))

	client.RemoteRef = &git.Ref{Name: "refs/heads/master"}
	cacheFile, err := client.prepareCacheDirectory("remote")
	assert.Nil(t, err)

	// Cache file should not exist
	fi, err := os.Stat(cacheFile)
	assert.True(t, os.IsNotExist(err))

	// Querying non-existing cache file will report nothing
	locks, err := client.SearchLocks(nil, 0, false, true)
	assert.NotNil(t, err)
	assert.Empty(t, locks)
	assert.Equal(t, 0, remoteQueries)

	// Need to include zero time in structure for equal to work
	zeroTime := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)

	// REMOTE QUERY: No cache file will be created when querying with a filter
	locks, err = client.SearchLocks(map[string]string{
		"key": "value",
	}, 0, false, false)
	assert.Nil(t, err)
	// Just make sure we have have received anything, content doesn't matter
	assert.Equal(t, 3, len(locks))
	assert.Equal(t, 1, remoteQueries)

	fi, err = os.Stat(cacheFile)
	assert.True(t, os.IsNotExist(err))

	// REMOTE QUERY: No cache file will be created when querying with a limit
	locks, err = client.SearchLocks(nil, 1, false, false)
	assert.Nil(t, err)
	// Just make sure we have have received anything, content doesn't matter
	assert.Equal(t, 1, len(locks))
	assert.Equal(t, 2, remoteQueries)

	fi, err = os.Stat(cacheFile)
	assert.True(t, os.IsNotExist(err))

	// REMOTE QUERY: locks will be reported and cache file should be created
	locks, err = client.SearchLocks(nil, 0, false, false)
	assert.Nil(t, err)
	assert.Equal(t, 3, remoteQueries)

	fi, err = os.Stat(cacheFile)
	assert.Nil(t, err)
	const size int64 = 300
	assert.Equal(t, size, fi.Size())

	expectedLocks := []Lock{
		Lock{Path: "folder/test1.dat", Id: "100", Owner: &User{Name: "Alice"}, LockedAt: zeroTime},
		Lock{Path: "folder/test2.dat", Id: "101", Owner: &User{Name: "Charles"}, LockedAt: zeroTime},
		Lock{Path: "folder/test3.dat", Id: "102", Owner: &User{Name: "Fred"}, LockedAt: zeroTime},
	}

	sort.Sort(LocksById(locks))
	assert.Equal(t, expectedLocks, locks)

	// Querying cache file should report same locks
	locks, err = client.SearchLocks(nil, 0, false, true)
	assert.Nil(t, err)
	assert.Equal(t, 3, remoteQueries)

	sort.Sort(LocksById(locks))
	assert.Equal(t, expectedLocks, locks)
}

func TestRefreshCache(t *testing.T) {
	var err error
	tempDir, err := ioutil.TempDir("", "testCacheLock")
	assert.Nil(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/locks/verify", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(lockVerifiableList{
			Theirs: []Lock{
				Lock{Id: "99", Path: "folder/test3.dat", Owner: &User{Name: "Alice"}},
				Lock{Id: "199", Path: "other/test1.dat", Owner: &User{Name: "Charles"}},
			},
			Ours: []Lock{
				Lock{Id: "101", Path: "folder/test1.dat", Owner: &User{Name: "Fred"}},
				Lock{Id: "102", Path: "folder/test2.dat", Owner: &User{Name: "Fred"}},
				Lock{Id: "103", Path: "root.dat", Owner: &User{Name: "Fred"}},
			},
		})
		assert.Nil(t, err)
	}))

	defer func() {
		srv.Close()
	}()

	lfsclient, err := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.url":    srv.URL + "/api",
		"user.name":  "Fred",
		"user.email": "fred@bloggs.com",
	}))
	require.Nil(t, err)

	client, err := NewClient("", lfsclient, config.New())
	assert.Nil(t, err)
	assert.Nil(t, client.SetupFileCache(tempDir))

	// Should start with no cached items
	locks, err := client.SearchLocks(nil, 0, true, false)
	assert.Nil(t, err)
	assert.Empty(t, locks)

	client.RemoteRef = &git.Ref{Name: "refs/heads/master"}
	_, _, err = client.SearchLocksVerifiable(100, false)
	assert.Nil(t, err)

	locks, err = client.SearchLocks(nil, 0, true, false)
	assert.Nil(t, err)
	// Need to include zero time in structure for equal to work
	zeroTime := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)

	// Sort locks for stable comparison
	sort.Sort(LocksById(locks))
	assert.Equal(t, []Lock{
		Lock{Path: "folder/test1.dat", Id: "101", Owner: &User{Name: "Fred"}, LockedAt: zeroTime},
		Lock{Path: "folder/test2.dat", Id: "102", Owner: &User{Name: "Fred"}, LockedAt: zeroTime},
		Lock{Path: "root.dat", Id: "103", Owner: &User{Name: "Fred"}, LockedAt: zeroTime},
		Lock{Path: "other/test1.dat", Id: "199", Owner: &User{Name: "Charles"}, LockedAt: zeroTime},
		Lock{Path: "folder/test3.dat", Id: "99", Owner: &User{Name: "Alice"}, LockedAt: zeroTime},
	}, locks)
}

func TestSearchLocksVerifiableWithCache(t *testing.T) {
	var err error
	tempDir, err := ioutil.TempDir("", "testCacheLock")
	assert.Nil(t, err)

	remoteQueries := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteQueries++

		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/locks/verify", r.URL.Path)

		body := lockVerifiableRequest{}
		if assert.Nil(t, json.NewDecoder(r.Body).Decode(&body)) {
			w.Header().Set("Content-Type", "application/json")
			list := lockVerifiableList{}
			if body.Cursor == "1" {
				list.Ours = []Lock{
					Lock{Path: "folder/1/test1.dat", Id: "111"},
				}
				list.Theirs = []Lock{
					Lock{Path: "folder/1/test2.dat", Id: "112"},
					Lock{Path: "folder/1/test3.dat", Id: "113"},
				}
			} else {
				list.Ours = []Lock{
					Lock{Path: "folder/0/test1.dat", Id: "101"},
					Lock{Path: "folder/0/test2.dat", Id: "102"},
				}
				list.Theirs = []Lock{
					Lock{Path: "folder/0/test3.dat", Id: "103"},
				}
				list.NextCursor = "1"
			}

			err := json.NewEncoder(w).Encode(&list)
			assert.Nil(t, err)
		} else {
			w.WriteHeader(500)
		}
	}))

	defer srv.Close()

	lfsclient, err := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.url":    srv.URL + "/api",
		"user.name":  "Fred",
		"user.email": "fred@bloggs.com",
	}))
	require.Nil(t, err)

	client, err := NewClient("", lfsclient, config.New())
	assert.Nil(t, client.SetupFileCache(tempDir))

	client.RemoteRef = &git.Ref{Name: "refs/heads/master"}
	cacheFile, err := client.prepareCacheDirectory("verifiable")
	assert.Nil(t, err)

	// Cache file should not exist
	fi, err := os.Stat(cacheFile)
	assert.True(t, os.IsNotExist(err))

	// Querying non-existing cache file will report nothing
	ourLocks, theirLocks, err := client.SearchLocksVerifiable(0, true)
	assert.NotNil(t, err)
	assert.Empty(t, ourLocks)
	assert.Empty(t, theirLocks)
	assert.Equal(t, 0, remoteQueries)

	// REMOTE QUERY: No cache file will be created when querying with a limit
	ourLocks, theirLocks, err = client.SearchLocksVerifiable(1, false)
	assert.Nil(t, err)
	// Just make sure we have have received anything, content doesn't matter
	assert.Equal(t, 1, len(ourLocks))
	assert.Equal(t, 0, len(theirLocks))
	assert.Equal(t, 1, remoteQueries)

	fi, err = os.Stat(cacheFile)
	assert.True(t, os.IsNotExist(err))

	// REMOTE QUERY: locks will be reported and cache file should be created
	ourLocks, theirLocks, err = client.SearchLocksVerifiable(0, false)
	assert.Nil(t, err)
	assert.Equal(t, 3, remoteQueries)

	fi, err = os.Stat(cacheFile)
	assert.Nil(t, err)
	const size int64 = 478
	assert.Equal(t, size, fi.Size())

	// Need to include zero time in structure for equal to work
	zeroTime := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)

	// Sort locks for stable comparison
	expectedOurLocks := []Lock{
		Lock{Path: "folder/0/test1.dat", Id: "101", LockedAt: zeroTime},
		Lock{Path: "folder/0/test2.dat", Id: "102", LockedAt: zeroTime},
		Lock{Path: "folder/1/test1.dat", Id: "111", LockedAt: zeroTime},
	}

	expectedTheirLocks := []Lock{
		Lock{Path: "folder/0/test3.dat", Id: "103", LockedAt: zeroTime},
		Lock{Path: "folder/1/test2.dat", Id: "112", LockedAt: zeroTime},
		Lock{Path: "folder/1/test3.dat", Id: "113", LockedAt: zeroTime},
	}

	sort.Sort(LocksById(ourLocks))
	assert.Equal(t, expectedOurLocks, ourLocks)
	sort.Sort(LocksById(theirLocks))
	assert.Equal(t, expectedTheirLocks, theirLocks)

	// Querying cache file should report same locks
	ourLocks, theirLocks, err = client.SearchLocksVerifiable(0, true)
	assert.Nil(t, err)
	assert.Equal(t, 3, remoteQueries)

	sort.Sort(LocksById(ourLocks))
	assert.Equal(t, expectedOurLocks, ourLocks)
	sort.Sort(LocksById(theirLocks))
	assert.Equal(t, expectedTheirLocks, theirLocks)
}
