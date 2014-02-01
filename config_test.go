package gitmedia

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestEndpointDefaultsToOrigin(t *testing.T) {
	config := &Configuration{
		map[string]string{"remote.origin.media": "abc"},
		[]string{},
	}

	assert.Equal(t, "abc", config.Endpoint())
}

func TestEndpointOverridesOrigin(t *testing.T) {
	config := &Configuration{
		map[string]string{
			"media.url":           "abc",
			"remote.origin.media": "def",
		},
		[]string{},
	}

	assert.Equal(t, "abc", config.Endpoint())
}
