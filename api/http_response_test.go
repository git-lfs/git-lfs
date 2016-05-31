package api_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/github/git-lfs/api"
	"github.com/stretchr/testify/assert"
)

func TestWrappedHttpResponsesMatchInternal(t *testing.T) {
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		Body:       ioutil.NopCloser(new(bytes.Buffer)),
	}
	wrapped := api.WrapHttpResponse(resp)

	assert.Equal(t, resp.Status, wrapped.Status())
	assert.Equal(t, resp.StatusCode, wrapped.StatusCode())
	assert.Equal(t, resp.Proto, wrapped.Proto())
	assert.Equal(t, resp.Body, wrapped.Body())
	assert.Equal(t, resp.Header, wrapped.Header())
}
