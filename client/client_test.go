package gitmediaclient

import (
	".."
	"github.com/bmizerany/assert"
	"testing"
)

func TestObjectUrl(t *testing.T) {
	oid := "oid"
	tests := map[string]string{
		"http://example.com":      "http://example.com/objects/oid",
		"http://example.com/":     "http://example.com/objects/oid",
		"http://example.com/foo":  "http://example.com/foo/objects/oid",
		"http://example.com/foo/": "http://example.com/foo/objects/oid",
	}

	config := gitmedia.Config()
	for endpoint, expected := range tests {
		config.Endpoint = endpoint
		assert.Equal(t, expected, ObjectUrl(oid).String())
	}
}
