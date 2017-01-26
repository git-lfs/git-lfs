package tq

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

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

		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "bar", r.Header.Get("Foo"))
		assert.Equal(t, "24", r.Header.Get("Content-Length"))
		assert.Equal(t, "application/vnd.git-lfs+json", r.Header.Get("Content-Type"))

		var tr Transfer
		assert.Nil(t, json.NewDecoder(r.Body).Decode(&tr))
		assert.Equal(t, "abc", tr.Oid)
		assert.EqualValues(t, 123, tr.Size)
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
