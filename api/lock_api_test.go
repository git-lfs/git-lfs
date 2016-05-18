package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/github/git-lfs/api"
	"github.com/stretchr/testify/assert"
)

var LockService api.LockService

func TestSuccessfullyObtainingALock(t *testing.T) {
	schema, resp := LockService.Lock(&api.LockRequest{
		Path:               "/path/to/file",
		LatestRemoteCommit: "deadbeef",
		Committer: api.Committer{
			Name:  "Jane Doe",
			Email: "jane@example.com",
		},
	})

	tc := &MethodTestCase{
		Schema:         schema,
		Response:       resp,
		ExpectedPath:   "/locks",
		ExpectedMethod: http.MethodPost,
		ExpectedResponse: &api.LockResponse{
			Lock: api.Lock{
				Id:   "some-lock-id",
				Path: "/path/to/file",
				Committer: api.Committer{
					Name:  "Jane Doe",
					Email: "jane@example.com",
				},
				CommitSHA: "deadbeef",
				LockedAt:  time.Date(2016, time.May, 18, 0, 0, 0, 0, time.UTC),
			},
		},
		Output: `
{
	"lock": {
		"id": "some-lock-id",
		"path": "/path/to/file",
		"committer": {
			"name": "Jane Doe",
			"email": "jane@example.com"
		},
		"commit_sha": "deadbeef",
		"locked_at": "2016-05-18T00:00:00Z"
	}
}`,
	}

	tc.Assert(t)
}

type MethodTestCase struct {
	Schema   *api.RequestSchema
	Response interface{}

	ExpectedPath     string
	ExpectedMethod   string
	ExpectedResponse interface{}

	Output string
}

func (c *MethodTestCase) Assert(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			called = true

			w.Write([]byte(c.Output))

			assert.Equal(t, c.ExpectedPath, req.URL.String())
			assert.Equal(t, c.ExpectedMethod, req.Method)
		},
	))

	client, _ := api.NewClient(server.URL)

	fmt.Println(client.Do(c.Schema))

	assert.Equal(t, true, called, "lfs/api: expected method %s to be called", c.ExpectedPath)
	assert.Equal(t, c.ExpectedResponse, c.Response)
}
