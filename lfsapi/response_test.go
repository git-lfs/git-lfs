package lfsapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/stretchr/testify/assert"
)

func TestAuthErrWithBody(t *testing.T) {
	var called uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/test" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		atomic.AddUint32(&called, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		w.Write([]byte(`{"message":"custom auth error"}`))
	}))
	defer srv.Close()

	req, err := http.NewRequest("GET", srv.URL+"/test", nil)
	assert.Nil(t, err)

	c, _ := NewClient(nil)
	_, err = c.Do(req)
	assert.NotNil(t, err)
	assert.True(t, errors.IsAuthError(err))
	assert.Equal(t, "Authentication required: custom auth error", err.Error())
	assert.EqualValues(t, 1, called)
}

func TestFatalWithBody(t *testing.T) {
	var called uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/test" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		atomic.AddUint32(&called, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"message":"custom fatal error"}`))
	}))
	defer srv.Close()

	req, err := http.NewRequest("GET", srv.URL+"/test", nil)
	assert.Nil(t, err)

	c, _ := NewClient(nil)
	_, err = c.Do(req)
	assert.NotNil(t, err)
	assert.True(t, errors.IsFatalError(err))
	assert.Equal(t, "Fatal error: custom fatal error", err.Error())
	assert.EqualValues(t, 1, called)
}

func TestWithNonFatal500WithBody(t *testing.T) {
	c, _ := NewClient(nil)

	var called uint32

	nonFatalCodes := map[int]string{
		501: "custom 501 error",
		507: "custom 507 error",
		509: "custom 509 error",
	}

	for nonFatalCode, expectedErr := range nonFatalCodes {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/test" {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			atomic.AddUint32(&called, 1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(nonFatalCode)
			w.Write([]byte(`{"message":"` + expectedErr + `"}`))
		}))

		req, err := http.NewRequest("GET", srv.URL+"/test", nil)
		assert.Nil(t, err)

		_, err = c.Do(req)
		t.Logf("non fatal code %d", nonFatalCode)
		assert.NotNil(t, err)
		assert.Equal(t, expectedErr, err.Error())
		srv.Close()
	}

	assert.EqualValues(t, 3, called)
}

func TestAuthErrWithoutBody(t *testing.T) {
	var called uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/test" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		atomic.AddUint32(&called, 1)
		w.WriteHeader(401)
	}))
	defer srv.Close()

	req, err := http.NewRequest("GET", srv.URL+"/test", nil)
	assert.Nil(t, err)

	c, _ := NewClient(nil)
	_, err = c.Do(req)
	assert.NotNil(t, err)
	assert.True(t, errors.IsAuthError(err))
	assert.True(t, strings.HasPrefix(err.Error(), "Authentication required: Authorization error:"), err.Error())
	assert.EqualValues(t, 1, called)
}

func TestFatalWithoutBody(t *testing.T) {
	var called uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/test" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		atomic.AddUint32(&called, 1)
		w.WriteHeader(500)
	}))
	defer srv.Close()

	req, err := http.NewRequest("GET", srv.URL+"/test", nil)
	assert.Nil(t, err)

	c, _ := NewClient(nil)
	_, err = c.Do(req)
	assert.NotNil(t, err)
	assert.True(t, errors.IsFatalError(err))
	assert.True(t, strings.HasPrefix(err.Error(), "Fatal error: Server error:"), err.Error())
	assert.EqualValues(t, 1, called)
}

func TestWithNonFatal500WithoutBody(t *testing.T) {
	c, _ := NewClient(nil)

	var called uint32

	nonFatalCodes := map[int]string{
		501: "Not Implemented:",
		507: "Insufficient server storage:",
		509: "Bandwidth limit exceeded:",
	}

	for nonFatalCode, errPrefix := range nonFatalCodes {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/test" {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			atomic.AddUint32(&called, 1)
			w.WriteHeader(nonFatalCode)
		}))

		req, err := http.NewRequest("GET", srv.URL+"/test", nil)
		assert.Nil(t, err)

		_, err = c.Do(req)
		t.Logf("non fatal code %d", nonFatalCode)
		assert.NotNil(t, err)
		assert.True(t, strings.HasPrefix(err.Error(), errPrefix))
		srv.Close()
	}

	assert.EqualValues(t, 3, called)
}
