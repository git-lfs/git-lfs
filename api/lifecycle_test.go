package api_test

import (
	"net/http"

	"github.com/github/git-lfs/api"
	"github.com/stretchr/testify/mock"
)

type MockLifecycle struct {
	mock.Mock
}

var _ api.Lifecycle = new(MockLifecycle)

func (l *MockLifecycle) Build(req *api.RequestSchema) (*http.Request, error) {
	args := l.Called(req)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*http.Request), args.Error(1)
}

func (l *MockLifecycle) Execute(req *http.Request, into interface{}) (api.Response, error) {
	args := l.Called(req, into)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(api.Response), args.Error(1)
}

func (l *MockLifecycle) Cleanup(resp api.Response) error {
	args := l.Called(resp)
	return args.Error(0)
}
