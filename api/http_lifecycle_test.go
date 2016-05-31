package api_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/github/git-lfs/api"
	"github.com/github/git-lfs/config"
	"github.com/stretchr/testify/assert"
)

type NopEndpointSource struct {
	Root string
}

func (e *NopEndpointSource) Endpoint(op string) config.Endpoint {
	return config.Endpoint{Url: e.Root}
}

var (
	source = &NopEndpointSource{"https://example.com"}
)

func TestHttpLifecycleMakesRequestsAgainstAbsolutePath(t *testing.T) {
	SetupTestCredentialsFunc()
	defer RestoreCredentialsFunc()

	l := api.NewHttpLifecycle(source)
	req, err := l.Build(&api.RequestSchema{
		Path:      "/foo",
		Operation: api.DownloadOperation,
	})

	assert.Nil(t, err)
	assert.Equal(t, "https://example.com/foo", req.URL.String())
}

func TestHttpLifecycleAttachesQueryParameters(t *testing.T) {
	SetupTestCredentialsFunc()
	defer RestoreCredentialsFunc()

	l := api.NewHttpLifecycle(source)
	req, err := l.Build(&api.RequestSchema{
		Path:      "/foo",
		Operation: api.DownloadOperation,
		Query: map[string]string{
			"a": "b",
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, "https://example.com/foo?a=b", req.URL.String())
}

func TestHttpLifecycleAttachesBodyWhenPresent(t *testing.T) {
	SetupTestCredentialsFunc()
	defer RestoreCredentialsFunc()

	l := api.NewHttpLifecycle(source)
	req, err := l.Build(&api.RequestSchema{
		Operation: api.DownloadOperation,
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
	SetupTestCredentialsFunc()
	defer RestoreCredentialsFunc()

	l := api.NewHttpLifecycle(source)
	req, err := l.Build(&api.RequestSchema{
		Operation: api.DownloadOperation,
	})

	assert.Nil(t, err)
	assert.Nil(t, req.Body)
}

func TestHttpLifecycleErrsWithoutOperation(t *testing.T) {
	SetupTestCredentialsFunc()
	defer RestoreCredentialsFunc()

	l := api.NewHttpLifecycle(source)
	req, err := l.Build(&api.RequestSchema{
		Path: "/foo",
	})

	assert.Equal(t, api.ErrNoOperationGiven, err)
	assert.Nil(t, req)
}

func TestHttpLifecycleExecutesRequestWithoutBody(t *testing.T) {
	SetupTestCredentialsFunc()
	defer RestoreCredentialsFunc()

	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		assert.Equal(t, "/path", r.URL.RequestURI())
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL+"/path", nil)

	l := api.NewHttpLifecycle(source)
	_, err := l.Execute(req, nil)

	assert.True(t, called)
	assert.Nil(t, err)
}

func TestHttpLifecycleExecutesRequestWithBody(t *testing.T) {
	SetupTestCredentialsFunc()
	defer RestoreCredentialsFunc()

	type Response struct {
		Foo string `json:"foo"`
	}

	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		w.Write([]byte("{\"foo\":\"bar\"}"))
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL+"/path", nil)

	l := api.NewHttpLifecycle(source)
	resp := new(Response)
	_, err := l.Execute(req, resp)

	assert.True(t, called)
	assert.Nil(t, err)
	assert.Equal(t, "bar", resp.Foo)
}
