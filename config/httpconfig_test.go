package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPConfig(t *testing.T) {
	c := NewHTTPConfig(EnvironmentOf(MapFetcher(map[string]string{
		"http.key":                         "root",
		"http.https://host.com.key":        "host",
		"http.https://user@host.com/a.key": "user-a",
		"http.https://user@host.com.key":   "user",
		"http.https://host.com/a.key":      "host-a",
		"http.https://host.com:8080.key":   "port",
	})))

	tests := map[string]string{
		"https://root.com/a/b/c":           "root",
		"https://host.com/":                "host",
		"https://host.com/a/b/c":           "host-a",
		"https://user:pass@host.com/a/b/c": "user-a",
		"https://user:pass@host.com/z/b/c": "user",
		"https://host.com:8080/a":          "port",
	}

	for rawurl, expected := range tests {
		value, _ := c.Get("http", "key", rawurl)
		assert.Equal(t, expected, value, rawurl)
	}
}
