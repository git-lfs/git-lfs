package locking

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/git-lfs/git-lfs/lfshttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

func TestAPILock(t *testing.T) {
	require.NotNil(t, createReqSchema)
	require.NotNil(t, createResSchema)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/locks" {
			w.WriteHeader(404)
			return
		}

		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, lfshttp.MediaType, r.Header.Get("Accept"))
		assert.Equal(t, lfshttp.MediaType, r.Header.Get("Content-Type"))
		assert.Equal(t, "53", r.Header.Get("Content-Length"))

		reqLoader, body := gojsonschema.NewReaderLoader(r.Body)
		lockReq := &lockRequest{}
		err := json.NewDecoder(body).Decode(lockReq)
		r.Body.Close()
		assert.Nil(t, err)
		assert.Equal(t, "refs/heads/master", lockReq.Ref.Name)
		assert.Equal(t, "request", lockReq.Path)
		assertSchema(t, createReqSchema, reqLoader)

		w.Header().Set("Content-Type", "application/json")
		resLoader, resWriter := gojsonschema.NewWriterLoader(w)
		err = json.NewEncoder(resWriter).Encode(&lockResponse{
			Lock: &Lock{
				Id:   "1",
				Path: "response",
			},
		})
		assert.Nil(t, err)
		assertSchema(t, createResSchema, resLoader)
	}))
	defer srv.Close()

	c, err := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.url": srv.URL + "/api",
	}))
	require.Nil(t, err)

	lc := &lockClient{Client: c}
	lockRes, res, err := lc.Lock("", &lockRequest{Path: "request", Ref: &lockRef{Name: "refs/heads/master"}})
	require.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "1", lockRes.Lock.Id)
	assert.Equal(t, "response", lockRes.Lock.Path)
}

func TestAPIUnlock(t *testing.T) {
	require.NotNil(t, delReqSchema)
	require.NotNil(t, createResSchema)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/locks/123/unlock" {
			w.WriteHeader(404)
			return
		}

		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, lfshttp.MediaType, r.Header.Get("Accept"))
		assert.Equal(t, lfshttp.MediaType, r.Header.Get("Content-Type"))

		reqLoader, body := gojsonschema.NewReaderLoader(r.Body)
		unlockReq := &unlockRequest{}
		err := json.NewDecoder(body).Decode(unlockReq)
		r.Body.Close()
		assert.Nil(t, err)
		assert.True(t, unlockReq.Force)
		assertSchema(t, delReqSchema, reqLoader)

		w.Header().Set("Content-Type", "application/json")
		resLoader, resWriter := gojsonschema.NewWriterLoader(w)
		err = json.NewEncoder(resWriter).Encode(&unlockResponse{
			Lock: &Lock{
				Id:   "123",
				Path: "response",
			},
		})
		assert.Nil(t, err)
		assertSchema(t, createResSchema, resLoader)
	}))
	defer srv.Close()

	c, err := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.url": srv.URL + "/api",
	}))
	require.Nil(t, err)

	lc := &lockClient{Client: c}
	unlockRes, res, err := lc.Unlock(&git.Ref{
		Name: "master",
		Sha:  "6161616161616161616161616161616161616161",
		Type: git.RefTypeLocalBranch,
	}, "", "123", true)
	require.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "123", unlockRes.Lock.Id)
	assert.Equal(t, "response", unlockRes.Lock.Path)
}

func TestAPISearch(t *testing.T) {
	require.NotNil(t, listResSchema)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/locks" {
			w.WriteHeader(404)
			return
		}

		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, lfshttp.MediaType, r.Header.Get("Accept"))
		assert.Equal(t, "", r.Header.Get("Content-Type"))

		q := r.URL.Query()
		assert.Equal(t, "A", q.Get("a"))
		assert.Equal(t, "cursor", q.Get("cursor"))
		assert.Equal(t, "5", q.Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		resLoader, resWriter := gojsonschema.NewWriterLoader(w)
		err := json.NewEncoder(resWriter).Encode(&lockList{
			Locks: []Lock{
				{Id: "1"},
				{Id: "2"},
			},
		})
		assert.Nil(t, err)
		assertSchema(t, listResSchema, resLoader)
	}))
	defer srv.Close()

	c, err := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
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

func TestAPISearchVerifiable(t *testing.T) {
	require.NotNil(t, verifyResSchema)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/locks/verify" {
			w.WriteHeader(404)
			return
		}

		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, lfshttp.MediaType, r.Header.Get("Accept"))
		assert.Equal(t, lfshttp.MediaType, r.Header.Get("Content-Type"))

		body := lockVerifiableRequest{}
		if assert.Nil(t, json.NewDecoder(r.Body).Decode(&body)) {
			assert.Equal(t, "cursor", body.Cursor)
			assert.Equal(t, 5, body.Limit)
		}

		w.Header().Set("Content-Type", "application/json")
		resLoader, resWriter := gojsonschema.NewWriterLoader(w)
		err := json.NewEncoder(resWriter).Encode(&lockVerifiableList{
			Ours: []Lock{
				{Id: "1"},
				{Id: "2"},
			},
			Theirs: []Lock{
				{Id: "3"},
			},
		})
		assert.Nil(t, err)
		assertSchema(t, verifyResSchema, resLoader)
	}))
	defer srv.Close()

	c, err := lfsapi.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.url": srv.URL + "/api",
	}))
	require.Nil(t, err)

	lc := &lockClient{Client: c}
	locks, res, err := lc.SearchVerifiable("", &lockVerifiableRequest{
		Cursor: "cursor",
		Limit:  5,
	})
	require.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, 2, len(locks.Ours))
	assert.Equal(t, "1", locks.Ours[0].Id)
	assert.Equal(t, "2", locks.Ours[1].Id)
	assert.Equal(t, 1, len(locks.Theirs))
	assert.Equal(t, "3", locks.Theirs[0].Id)
}

var (
	createReqSchema *sourcedSchema
	createResSchema *sourcedSchema
	delReqSchema    *sourcedSchema
	listResSchema   *sourcedSchema
	verifyResSchema *sourcedSchema
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("getwd error:", err)
		return
	}

	createReqSchema = getSchema(wd, "schemas/http-lock-create-request-schema.json")
	createResSchema = getSchema(wd, "schemas/http-lock-create-response-schema.json")
	delReqSchema = getSchema(wd, "schemas/http-lock-delete-request-schema.json")
	listResSchema = getSchema(wd, "schemas/http-lock-list-response-schema.json")
	verifyResSchema = getSchema(wd, "schemas/http-lock-verify-response-schema.json")
}

type sourcedSchema struct {
	Source string
	*gojsonschema.Schema
}

func getSchema(wd, relpath string) *sourcedSchema {
	abspath := filepath.ToSlash(filepath.Join(wd, relpath))
	s, err := gojsonschema.NewSchema(gojsonschema.NewReferenceLoader(fmt.Sprintf("file:///%s", abspath)))
	if err != nil {
		fmt.Printf("schema load error for %q: %+v\n", relpath, err)
	}
	return &sourcedSchema{Source: relpath, Schema: s}
}

func assertSchema(t *testing.T, schema *sourcedSchema, dataLoader gojsonschema.JSONLoader) {
	res, err := schema.Validate(dataLoader)
	if assert.Nil(t, err) {
		if res.Valid() {
			return
		}

		resErrors := res.Errors()
		valErrors := make([]string, 0, len(resErrors))
		for _, resErr := range resErrors {
			valErrors = append(valErrors, resErr.String())
		}
		t.Errorf("Schema: %s\n%s", schema.Source, strings.Join(valErrors, "\n"))
	}
}
