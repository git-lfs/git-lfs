package api_test

import (
	"net/http"
	"testing"

	"github.com/github/git-lfs/api"
)

var LockService api.LockService

func TestSuccessfullyObtainingALock(t *testing.T) {
	got, body := LockService.Lock(new(api.LockRequest))

	AssertSchema(t, &api.RequestSchema{
		Method: http.MethodPost,
		Path:   "/locks",
		Body:   new(api.LockRequest),
		Into:   body,
	}, got)
}
