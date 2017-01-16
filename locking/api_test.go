package locking

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPILock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/locks" {
			w.WriteHeader(404)
			return
		}

		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, lfsapi.MediaType, r.Header.Get("Accept"))
		assert.Equal(t, lfsapi.MediaType, r.Header.Get("Content-Type"))

		lockReq := &lockRequest{}
		err := json.NewDecoder(r.Body).Decode(lockReq)
		r.Body.Close()
		assert.Nil(t, err)
		assert.Equal(t, "request", lockReq.Path)

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(&lockResponse{
			Lock: &Lock{
				Id:   "1",
				Path: "response",
			},
		})
		assert.Nil(t, err)
	}))
	defer srv.Close()

	c, err := lfsapi.NewClient(nil, lfsapi.TestEnv(map[string]string{
		"lfs.url": srv.URL + "/api",
	}))
	require.Nil(t, err)

	lc := &lockClient{Client: c}
	lockRes, res, err := lc.Lock("", &lockRequest{Path: "request"})
	require.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "1", lockRes.Lock.Id)
	assert.Equal(t, "response", lockRes.Lock.Path)
}

func TestAPIUnlock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/locks/123/unlock" {
			w.WriteHeader(404)
			return
		}

		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, lfsapi.MediaType, r.Header.Get("Accept"))
		assert.Equal(t, lfsapi.MediaType, r.Header.Get("Content-Type"))

		unlockReq := &unlockRequest{}
		err := json.NewDecoder(r.Body).Decode(unlockReq)
		r.Body.Close()
		assert.Nil(t, err)
		assert.Equal(t, "123", unlockReq.Id)
		assert.True(t, unlockReq.Force)

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(&unlockResponse{
			Lock: &Lock{
				Id:   "123",
				Path: "response",
			},
		})
		assert.Nil(t, err)
	}))
	defer srv.Close()

	c, err := lfsapi.NewClient(nil, lfsapi.TestEnv(map[string]string{
		"lfs.url": srv.URL + "/api",
	}))
	require.Nil(t, err)

	lc := &lockClient{Client: c}
	unlockRes, res, err := lc.Unlock("", "123", true)
	require.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "123", unlockRes.Lock.Id)
	assert.Equal(t, "response", unlockRes.Lock.Path)
}

func TestAPISearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/locks" {
			w.WriteHeader(404)
			return
		}

		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, lfsapi.MediaType, r.Header.Get("Accept"))
		assert.Equal(t, "", r.Header.Get("Content-Type"))

		q := r.URL.Query()
		assert.Equal(t, "A", q.Get("a"))
		assert.Equal(t, "cursor", q.Get("cursor"))
		assert.Equal(t, "5", q.Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(&lockList{
			Locks: []Lock{
				{Id: "1"},
				{Id: "2"},
			},
		})
		assert.Nil(t, err)
	}))
	defer srv.Close()

	c, err := lfsapi.NewClient(nil, lfsapi.TestEnv(map[string]string{
		"lfs.url": srv.URL + "/api",
	}))
	require.Nil(t, err)

	lc := &lockClient{Client: c}
	locks, res, err := lc.Search("", &lockSearchRequest{
		Filters: []lockFilter{
			{Property: "a", Value: "A"},
		},
		Cursor: "cursor",
		Limit:  5,
	})
	require.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, 2, len(locks.Locks))
	assert.Equal(t, "1", locks.Locks[0].Id)
	assert.Equal(t, "2", locks.Locks[1].Id)
}
