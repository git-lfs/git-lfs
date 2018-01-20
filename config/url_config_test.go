package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLConfig(t *testing.T) {
	u := NewURLConfig(EnvironmentOf(MapFetcher(map[string][]string{
		"http.key":                           []string{"root", "root-2"},
		"http.https://host.com.key":          []string{"host", "host-2"},
		"http.https://user@host.com/a.key":   []string{"user-a", "user-b"},
		"http.https://user@host.com.key":     []string{"user", "user-2"},
		"http.https://host.com/a.key":        []string{"host-a", "host-b"},
		"http.https://host.com:8080.key":     []string{"port", "port-2"},
		"http.https://host.com/repo.git.key": []string{".git"},
		"http.https://host.com/repo.key":     []string{"no .git"},
		"http.https://host.com/repo2.key":    []string{"no .git"},
	})))

	getOne := map[string]string{
		"https://root.com/a/b/c":                      "root-2",
		"https://host.com/":                           "host-2",
		"https://host.com/a/b/c":                      "host-b",
		"https://user:pass@host.com/a/b/c":            "user-b",
		"https://user:pass@host.com/z/b/c":            "user-2",
		"https://host.com:8080/a":                     "port-2",
		"https://host.com/repo.git/info/lfs":          ".git",
		"https://host.com/repo.git/info":              ".git",
		"https://host.com/repo.git":                   ".git",
		"https://host.com/repo":                       "no .git",
		"https://host.com/repo2.git/info/lfs/foo/bar": "no .git",
		"https://host.com/repo2.git/info/lfs":         "no .git",
		"https://host.com/repo2.git/info":             "host-2", // doesn't match /.git/info/lfs\Z/
		"https://host.com/repo2.git":                  "host-2", // ditto
		"https://host.com/repo2":                      "no .git",
	}

	for rawurl, expected := range getOne {
		value, _ := u.Get("http", rawurl, "key")
		assert.Equal(t, expected, value, "get one: "+rawurl)
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
		values := u.GetAll("http", rawurl, "key")
		assert.Equal(t, expected, values, "get all: "+rawurl)
	}
}
