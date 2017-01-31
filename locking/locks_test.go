package locking

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type LocksById []Lock

func (a LocksById) Len() int           { return len(a) }
func (a LocksById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a LocksById) Less(i, j int) bool { return a[i].Id < a[j].Id }

func TestRefreshCache(t *testing.T) {
	var err error
	tempDir, err := ioutil.TempDir("", "testCacheLock")
	assert.Nil(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/locks", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(lockList{
			Locks: []Lock{
				Lock{Id: "99", Path: "folder/test3.dat", Committer: &Committer{Name: "Alice", Email: "alice@wonderland.com"}},
				Lock{Id: "101", Path: "folder/test1.dat", Committer: &Committer{Name: "Fred", Email: "fred@bloggs.com"}},
				Lock{Id: "102", Path: "folder/test2.dat", Committer: &Committer{Name: "Fred", Email: "fred@bloggs.com"}},
				Lock{Id: "103", Path: "root.dat", Committer: &Committer{Name: "Fred", Email: "fred@bloggs.com"}},
				Lock{Id: "199", Path: "other/test1.dat", Committer: &Committer{Name: "Charles", Email: "charles@incharge.com"}},
			},
		})
		assert.Nil(t, err)
	}))

	defer func() {
		srv.Close()
	}()

	lfsclient, err := lfsapi.NewClient(nil, lfsapi.TestEnv(map[string]string{
		"lfs.url":    srv.URL + "/api",
		"user.name":  "Fred",
		"user.email": "fred@bloggs.com",
	}))
	require.Nil(t, err)

	client, err := NewClient("", lfsclient)
	assert.Nil(t, err)
	assert.Nil(t, client.SetupFileCache(tempDir))

	// Should start with no cached items
	locks, err := client.SearchLocks(nil, 0, true)
	assert.Nil(t, err)
	assert.Empty(t, locks)

	// Should load from test data, just Fred's
	err = client.refreshLockCache()
	assert.Nil(t, err)

	locks, err = client.SearchLocks(nil, 0, true)
	assert.Nil(t, err)
	// Need to include zero time in structure for equal to work
	zeroTime := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)

	// Sort locks for stable comparison
	sort.Sort(LocksById(locks))
	assert.Equal(t, []Lock{
		Lock{Path: "folder/test1.dat", Id: "101", Committer: &Committer{Name: "Fred", Email: "fred@bloggs.com"}, LockedAt: zeroTime},
		Lock{Path: "folder/test2.dat", Id: "102", Committer: &Committer{Name: "Fred", Email: "fred@bloggs.com"}, LockedAt: zeroTime},
		Lock{Path: "root.dat", Id: "103", Committer: &Committer{Name: "Fred", Email: "fred@bloggs.com"}, LockedAt: zeroTime},
	}, locks)
}

func TestGetVerifiableLocks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	lfsclient, err := lfsapi.NewClient(nil, lfsapi.TestEnv(map[string]string{
		"lfs.url":    srv.URL + "/api",
		"user.name":  "Fred",
		"user.email": "fred@bloggs.com",
	}))
	require.Nil(t, err)

	client, err := NewClient("", lfsclient)
	assert.Nil(t, err)

	ourLocks, theirLocks, err := client.VerifiableLocks(0)
	assert.Nil(t, err)

	// Need to include zero time in structure for equal to work
	zeroTime := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)

	// Sort locks for stable comparison
	sort.Sort(LocksById(ourLocks))
	assert.Equal(t, []Lock{
		Lock{Path: "folder/0/test1.dat", Id: "101", LockedAt: zeroTime},
		Lock{Path: "folder/0/test2.dat", Id: "102", LockedAt: zeroTime},
		Lock{Path: "folder/1/test1.dat", Id: "111", LockedAt: zeroTime},
	}, ourLocks)

	sort.Sort(LocksById(theirLocks))
	assert.Equal(t, []Lock{
		Lock{Path: "folder/0/test3.dat", Id: "103", LockedAt: zeroTime},
		Lock{Path: "folder/1/test2.dat", Id: "112", LockedAt: zeroTime},
		Lock{Path: "folder/1/test3.dat", Id: "113", LockedAt: zeroTime},
	}, theirLocks)
}
