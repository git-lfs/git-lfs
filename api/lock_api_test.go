package api_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/github/git-lfs/api"
	"github.com/stretchr/testify/assert"
)

var LockService api.LockService

func TestSuccessfullyObtainingALock(t *testing.T) {
	now := time.Now()

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
		ExpectedPath:   "/locks",
		ExpectedMethod: http.MethodPost,
		Resp:           resp,
		Output: &api.LockResponse{
			Lock: api.Lock{
				Id:   "some-lock-id",
				Path: "/path/to/file",
				Committer: api.Committer{
					Name:  "Jane Doe",
					Email: "jane@example.com",
				},
				CommitSHA: "deadbeef",
				LockedAt:  now,
			},
		},
	}

	tc.Assert(t)

	assert.Nil(t, resp.Err)
	assert.Equal(t, "", resp.CommitNeeded)
	assert.Equal(t, "some-lock-id", resp.Lock.Id)
	assert.Equal(t, "/path/to/file", resp.Lock.Path)
	assert.Equal(t, api.Committer{"Jane Doe", "jane@example.com"}, resp.Lock.Committer)
	assert.Equal(t, now, resp.Lock.LockedAt)
	assert.True(t, resp.Lock.Active())
}
