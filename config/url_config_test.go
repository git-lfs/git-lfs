package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLConfig(t *testing.T) {
	u := NewURLConfig(EnvironmentOf(MapFetcher(map[string][]string{
		"http.key":                         []string{"root", "root-2"},
		"http.https://host.com.key":        []string{"host", "host-2"},
		"http.https://user@host.com/a.key": []string{"user-a", "user-b"},
		"http.https://user@host.com.key":   []string{"user", "user-2"},
		"http.https://host.com/a.key":      []string{"host-a", "host-b"},
		"http.https://host.com:8080.key":   []string{"port", "port-2"},
	})))

	getOne := map[string]string{
		"https://root.com/a/b/c":           "root-2",
		"https://host.com/":                "host-2",
		"https://host.com/a/b/c":           "host-b",
		"https://user:pass@host.com/a/b/c": "user-b",
		"https://user:pass@host.com/z/b/c": "user-2",
		"https://host.com:8080/a":          "port-2",
	}

	for rawurl, expected := range getOne {
		value, _ := u.Get("http", "key", rawurl)
		assert.Equal(t, expected, value, rawurl)
	}

	getAll := map[string][]string{
		"https://root.com/a/b/c":           []string{"root", "root-2"},
		"https://host.com/":                []string{"host", "host-2"},
		"https://host.com/a/b/c":           []string{"host-a", "host-b"},
		"https://user:pass@host.com/a/b/c": []string{"user-a", "user-b"},
		"https://user:pass@host.com/z/b/c": []string{"user", "user-2"},
		"https://host.com:8080/a":          []string{"port", "port-2"},
	}

	for rawurl, expected := range getAll {
		values := u.GetAll("http", "key", rawurl)
		assert.Equal(t, expected, values, rawurl)
	}
}
