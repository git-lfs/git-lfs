package tq

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/stretchr/testify/assert"
)

func TestVerifyWithoutAction(t *testing.T) {
	c := &lfsapi.Client{}
	tr := &Transfer{
		Oid:  "abc",
		Size: 123,
	}

	assert.Nil(t, verifyUpload(c, tr))
}

func TestVerifySuccess(t *testing.T) {
	var called uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/verify" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		atomic.AddUint32(&called, 1)

		var tr Transfer
		err := json.NewDecoder(r.Body).Decode(&tr)

		assert.Nil(t, err)
		assert.Equal(t, "abc", tr.Oid)
		assert.EqualValues(t, 123, tr.Size)
		assert.Equal(t, "bar", r.Header.Get("Foo"))
		assert.Equal(t, "application/vnd.git-lfs+json", r.Header.Get("Content-Type"))
		assert.Equal(t, "24", r.Header.Get("Content-Length"))
	}))
	defer srv.Close()

	c := &lfsapi.Client{}
	tr := &Transfer{
		Oid:  "abc",
		Size: 123,
		Actions: map[string]*Action{
			"verify": &Action{
				Href: srv.URL + "/verify",
				Header: map[string]string{
					"foo": "bar",
				},
			},
		},
	}

	assert.Nil(t, verifyUpload(c, tr))
	assert.EqualValues(t, 1, called)
}

func TestVerifyAuthErrWithBody(t *testing.T) {
	var called uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/verify" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		atomic.AddUint32(&called, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		w.Write([]byte(`{"message":"custom auth error"}`))
	}))
	defer srv.Close()

	c := &lfsapi.Client{}
	tr := &Transfer{
		Oid:  "abc",
		Size: 123,
		Actions: map[string]*Action{
			"verify": &Action{
				Href: srv.URL + "/verify",
			},
		},
	}

	err := verifyUpload(c, tr)
	assert.NotNil(t, err)
	assert.True(t, errors.IsAuthError(err))
	assert.Equal(t, "Authentication required: http: custom auth error", err.Error())
	assert.EqualValues(t, 1, called)
}

func TestVerifyFatalWithBody(t *testing.T) {
	var called uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/verify" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		atomic.AddUint32(&called, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"message":"custom fatal error"}`))
	}))
	defer srv.Close()

	c := &lfsapi.Client{}
	tr := &Transfer{
		Oid:  "abc",
		Size: 123,
		Actions: map[string]*Action{
			"verify": &Action{
				Href: srv.URL + "/verify",
			},
		},
	}

	err := verifyUpload(c, tr)
	assert.NotNil(t, err)
	assert.True(t, errors.IsFatalError(err))
	assert.Equal(t, "Fatal error: http: custom fatal error", err.Error())
	assert.EqualValues(t, 1, called)
}

func TestVerifyWithNonFatal500WithBody(t *testing.T) {
	c := &lfsapi.Client{}
	tr := &Transfer{
		Oid:  "abc",
		Size: 123,
		Actions: map[string]*Action{
			"verify": &Action{},
		},
	}

	var called uint32

	nonFatalCodes := map[int]string{
		501: "custom 501 error",
		507: "custom 507 error",
		509: "custom 509 error",
	}

	for nonFatalCode, expectedErr := range nonFatalCodes {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/verify" {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			atomic.AddUint32(&called, 1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(nonFatalCode)
			w.Write([]byte(`{"message":"` + expectedErr + `"}`))
		}))

		tr.Actions["verify"].Href = srv.URL + "/verify"
		err := verifyUpload(c, tr)
		t.Logf("non fatal code %d", nonFatalCode)
		assert.NotNil(t, err)
		assert.Equal(t, "http: "+expectedErr, err.Error())
		srv.Close()
	}

	assert.EqualValues(t, 3, called)
}

func TestVerifyAuthErrWithoutBody(t *testing.T) {
	var called uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/verify" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		atomic.AddUint32(&called, 1)
		w.WriteHeader(401)
	}))
	defer srv.Close()

	c := &lfsapi.Client{}
	tr := &Transfer{
		Oid:  "abc",
		Size: 123,
		Actions: map[string]*Action{
			"verify": &Action{
				Href: srv.URL + "/verify",
			},
		},
	}

	err := verifyUpload(c, tr)
	assert.NotNil(t, err)
	assert.True(t, errors.IsAuthError(err))
	assert.True(t, strings.HasPrefix(err.Error(), "Authentication required: Authorization error:"), err.Error())
	assert.EqualValues(t, 1, called)
}

func TestVerifyFatalWithoutBody(t *testing.T) {
	var called uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/verify" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		atomic.AddUint32(&called, 1)
		w.WriteHeader(500)
	}))
	defer srv.Close()

	c := &lfsapi.Client{}
	tr := &Transfer{
		Oid:  "abc",
		Size: 123,
		Actions: map[string]*Action{
			"verify": &Action{
				Href: srv.URL + "/verify",
			},
		},
	}

	err := verifyUpload(c, tr)
	assert.NotNil(t, err)
	assert.True(t, errors.IsFatalError(err))
	assert.True(t, strings.HasPrefix(err.Error(), "Fatal error: Server error:"), err.Error())
	assert.EqualValues(t, 1, called)
}

func TestVerifyWithNonFatal500WithoutBody(t *testing.T) {
	c := &lfsapi.Client{}
	tr := &Transfer{
		Oid:  "abc",
		Size: 123,
		Actions: map[string]*Action{
			"verify": &Action{},
		},
	}

	var called uint32

	nonFatalCodes := map[int]string{
		501: "Not Implemented:",
		507: "Insufficient server storage:",
		509: "Bandwidth limit exceeded:",
	}

	for nonFatalCode, errPrefix := range nonFatalCodes {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != "/verify" {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			atomic.AddUint32(&called, 1)
			w.WriteHeader(nonFatalCode)
		}))

		tr.Actions["verify"].Href = srv.URL + "/verify"
		err := verifyUpload(c, tr)
		t.Logf("non fatal code %d", nonFatalCode)
		assert.NotNil(t, err)
		assert.True(t, strings.HasPrefix(err.Error(), errPrefix))
		srv.Close()
	}

	assert.EqualValues(t, 3, called)
}
