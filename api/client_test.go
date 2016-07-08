package api_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/github/git-lfs/api"
	"github.com/stretchr/testify/assert"
)

func TestClientUsesLifecycleToExecuteSchemas(t *testing.T) {
	schema := new(api.RequestSchema)
	req := new(http.Request)
	resp := new(api.HttpResponse)

	lifecycle := new(MockLifecycle)
	lifecycle.On("Build", schema).Return(req, nil).Once()
	lifecycle.On("Execute", req, schema.Into).Return(resp, nil).Once()
	lifecycle.On("Cleanup", resp).Return(nil).Once()

	client := api.NewClient(lifecycle)
	r1, err := client.Do(schema)

	assert.Equal(t, resp, r1)
	assert.Nil(t, err)
	lifecycle.AssertExpectations(t)
}

func TestClientHaltsIfSchemaCannotBeBuilt(t *testing.T) {
	schema := new(api.RequestSchema)

	lifecycle := new(MockLifecycle)
	lifecycle.On("Build", schema).Return(nil, errors.New("uh-oh!")).Once()

	client := api.NewClient(lifecycle)
	resp, err := client.Do(schema)

	lifecycle.AssertExpectations(t)
	assert.Nil(t, resp)
	assert.Equal(t, "uh-oh!", err.Error())
}

func TestClientHaltsIfSchemaCannotBeExecuted(t *testing.T) {
	schema := new(api.RequestSchema)
	req := new(http.Request)

	lifecycle := new(MockLifecycle)
	lifecycle.On("Build", schema).Return(req, nil).Once()
	lifecycle.On("Execute", req, schema.Into).Return(nil, errors.New("uh-oh!")).Once()

	client := api.NewClient(lifecycle)
	resp, err := client.Do(schema)

	lifecycle.AssertExpectations(t)
	assert.Nil(t, resp)
	assert.Equal(t, "uh-oh!", err.Error())
}

func TestClientReturnsCleanupErrors(t *testing.T) {
	schema := new(api.RequestSchema)
	req := new(http.Request)
	resp := new(api.HttpResponse)

	lifecycle := new(MockLifecycle)
	lifecycle.On("Build", schema).Return(req, nil).Once()
	lifecycle.On("Execute", req, schema.Into).Return(resp, nil).Once()
	lifecycle.On("Cleanup", resp).Return(errors.New("uh-oh!")).Once()

	client := api.NewClient(lifecycle)
	r1, err := client.Do(schema)

	lifecycle.AssertExpectations(t)
	assert.Nil(t, r1)
	assert.Equal(t, "uh-oh!", err.Error())
}
