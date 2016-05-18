package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/github/git-lfs/api"
	"github.com/stretchr/testify/assert"
)

type MethodTestCase struct {
	Schema *api.RequestSchema

	ExpectedPath   string
	ExpectedMethod string

	Output interface{}
}

func (c *MethodTestCase) Assert(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			called = true

			if err := json.NewEncoder(w).Encode(c.Output); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, c.ExpectedPath, req.URL.String())
			assert.Equal(t, c.ExpectedMethod, req.Method)
		},
	))

	client, _ := api.NewClient(server.URL)

	_, err := client.Do(c.Schema)

	assert.Nil(t, err)
	assert.Equal(t, true, called, "lfs/api: expected method %s to be called", c.ExpectedPath)
}
