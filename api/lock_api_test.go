package api_test

import (
	"testing"
	"time"

	"github.com/git-lfs/git-lfs/api"
	"github.com/git-lfs/git-lfs/api/schema"
)

var LockService api.LockService

func TestSuccessfullyObtainingALock(t *testing.T) {
	got, body := LockService.Lock(new(api.LockRequest))

	AssertRequestSchema(t, &api.RequestSchema{
		Method:    "POST",
		Path:      "/locks",
		Operation: api.UploadOperation,
		Body:      new(api.LockRequest),
		Into:      body,
	}, got)
}

func TestLockSearchWithFilters(t *testing.T) {
	got, body := LockService.Search(&api.LockSearchRequest{
		Filters: []api.Filter{
			{"branch", "master"},
			{"path", "/path/to/file"},
		},
	})

	AssertRequestSchema(t, &api.RequestSchema{
		Method: "GET",
		Query: map[string]string{
			"branch": "master",
			"path":   "/path/to/file",
		},
		Path:      "/locks",
		Operation: api.DownloadOperation,
		Into:      body,
	}, got)
}

func TestLockSearchWithNextCursor(t *testing.T) {
	got, body := LockService.Search(&api.LockSearchRequest{
		Cursor: "some-lock-id",
	})

	AssertRequestSchema(t, &api.RequestSchema{
		Method: "GET",
		Query: map[string]string{
			"cursor": "some-lock-id",
		},
		Path:      "/locks",
		Operation: api.DownloadOperation,
		Into:      body,
	}, got)
}

func TestLockSearchWithLimit(t *testing.T) {
	got, body := LockService.Search(&api.LockSearchRequest{
		Limit: 20,
	})

	AssertRequestSchema(t, &api.RequestSchema{
		Method: "GET",
		Query: map[string]string{
			"limit": "20",
		},
		Path:      "/locks",
		Operation: api.DownloadOperation,
		Into:      body,
	}, got)
}

func TestUnlockingALock(t *testing.T) {
	got, body := LockService.Unlock("some-lock-id", true)

	AssertRequestSchema(t, &api.RequestSchema{
		Method:    "POST",
		Path:      "/locks/some-lock-id/unlock",
		Operation: api.UploadOperation,
		Body: &api.UnlockRequest{
			Id:    "some-lock-id",
			Force: true,
		},
		Into: body,
	}, got)
}

func TestLockRequest(t *testing.T) {
	schema.Validate(t, schema.LockRequestSchema, &api.LockRequest{
		Path:               "/path/to/lock",
		LatestRemoteCommit: "deadbeef",
		Committer: api.Committer{
			Name:  "Jane Doe",
			Email: "jane@example.com",
		},
	})
}

func TestLockResponseWithLockedLock(t *testing.T) {
	schema.Validate(t, schema.LockResponseSchema, &api.LockResponse{
		Lock: &api.Lock{
			Id:   "some-lock-id",
			Path: "/lock/path",
			Committer: api.Committer{
				Name:  "Jane Doe",
				Email: "jane@example.com",
			},
			LockedAt: time.Now(),
		},
	})
}

func TestLockResponseWithUnlockedLock(t *testing.T) {
	schema.Validate(t, schema.LockResponseSchema, &api.LockResponse{
		Lock: &api.Lock{
			Id:   "some-lock-id",
			Path: "/lock/path",
			Committer: api.Committer{
				Name:  "Jane Doe",
				Email: "jane@example.com",
			},
			LockedAt:   time.Now(),
			UnlockedAt: time.Now(),
		},
	})
}

func TestLockResponseWithError(t *testing.T) {
	schema.Validate(t, schema.LockResponseSchema, &api.LockResponse{
		Err: "some error",
	})
}

func TestLockResponseWithCommitNeeded(t *testing.T) {
	schema.Validate(t, schema.LockResponseSchema, &api.LockResponse{
		CommitNeeded: "deadbeef",
	})
}

func TestLockResponseInvalidWithCommitAndError(t *testing.T) {
	schema.Refute(t, schema.LockResponseSchema, &api.LockResponse{
		Err:          "some error",
		CommitNeeded: "deadbeef",
	})
}

func TestUnlockRequest(t *testing.T) {
	schema.Validate(t, schema.UnlockRequestSchema, &api.UnlockRequest{
		Id:    "some-lock-id",
		Force: false,
	})
}

func TestUnlockResponseWithLock(t *testing.T) {
	schema.Validate(t, schema.UnlockResponseSchema, &api.UnlockResponse{
		Lock: &api.Lock{
			Id: "some-lock-id",
		},
	})
}

func TestUnlockResponseWithError(t *testing.T) {
	schema.Validate(t, schema.UnlockResponseSchema, &api.UnlockResponse{
		Err: "some-error",
	})
}

func TestUnlockResponseDoesNotAllowLockAndError(t *testing.T) {
	schema.Refute(t, schema.UnlockResponseSchema, &api.UnlockResponse{
		Lock: &api.Lock{
			Id: "some-lock-id",
		},
		Err: "some-error",
	})
}

func TestLockListWithLocks(t *testing.T) {
	schema.Validate(t, schema.LockListSchema, &api.LockList{
		Locks: []api.Lock{
			api.Lock{Id: "foo"},
			api.Lock{Id: "bar"},
		},
	})
}

func TestLockListWithNoResults(t *testing.T) {
	schema.Validate(t, schema.LockListSchema, &api.LockList{
		Locks: []api.Lock{},
	})
}

func TestLockListWithNextCursor(t *testing.T) {
	schema.Validate(t, schema.LockListSchema, &api.LockList{
		Locks: []api.Lock{
			api.Lock{Id: "foo"},
			api.Lock{Id: "bar"},
		},
		NextCursor: "baz",
	})
}

func TestLockListWithError(t *testing.T) {
	schema.Validate(t, schema.LockListSchema, &api.LockList{
		Err: "some error",
	})
}

func TestLockListWithErrorAndLocks(t *testing.T) {
	schema.Refute(t, schema.LockListSchema, &api.LockList{
		Locks: []api.Lock{
			api.Lock{Id: "foo"},
			api.Lock{Id: "bar"},
		},
		Err: "this isn't possible!",
	})
}
