package api_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/github/git-lfs/api"
	"github.com/github/git-lfs/vendor/_nuts/github.com/stretchr/testify/assert"
)

var (
	root, _ = url.Parse("https://example.com")
)

func TestHttpLifecycleMakesRequestsAgainstAbsolutePath(t *testing.T) {
	l := api.NewHttpLifecycle(root)
	req, err := l.Build(&api.RequestSchema{
		Path: "/foo",
	})

	assert.Nil(t, err)
	assert.Equal(t, "https://example.com/foo", req.URL.String())
}

func TestHttpLifecycleAttachesQueryParameters(t *testing.T) {
	l := api.NewHttpLifecycle(root)
	req, err := l.Build(&api.RequestSchema{
		Path: "/foo",
		Query: map[string]string{
			"a": "b",
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, "https://example.com/foo?a=b", req.URL.String())
}

func TestHttpLifecycleAttachesBodyWhenPresent(t *testing.T) {
	l := api.NewHttpLifecycle(root)
	req, err := l.Build(&api.RequestSchema{
		Body: struct {
			Foo string `json:"foo"`
		}{"bar"},
	})

	assert.Nil(t, err)

	body, err := ioutil.ReadAll(req.Body)
	assert.Nil(t, err)
	assert.Equal(t, "{\"foo\":\"bar\"}", string(body))
}

func TestHttpLifecycleDoesNotAttachBodyWhenEmpty(t *testing.T) {
	l := api.NewHttpLifecycle(root)
	req, err := l.Build(&api.RequestSchema{})

	assert.Nil(t, err)
	assert.Nil(t, req.Body)
}

func TestHttpLifecycleExecutesRequestWithoutBody(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		assert.Equal(t, "/path", r.URL.RequestURI())
	}))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/path", nil)

	l := api.NewHttpLifecycle(nil)
	_, err := l.Execute(req, nil)

	assert.True(t, called)
	assert.Nil(t, err)
}

func TestHttpLifecycleExecutesRequestWithBody(t *testing.T) {
	type Response struct {
		Foo string `json:"foo"`
	}

	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		w.Write([]byte("{\"foo\":\"bar\"}"))
	}))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/path", nil)

	l := api.NewHttpLifecycle(nil)
	resp := new(Response)
	_, err := l.Execute(req, resp)

	assert.True(t, called)
	assert.Nil(t, err)
	assert.Equal(t, "bar", resp.Foo)
}
