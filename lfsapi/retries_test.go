package lfsapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithRetries(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req = WithRetries(req, 1)

	n, ok := Retries(req)
	assert.True(t, ok)
	assert.Equal(t, 1, n)
}

func TestRetriesOnUnannotatedRequest(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)

	n, ok := Retries(req)
	assert.False(t, ok)
	assert.Equal(t, 0, n)
}
