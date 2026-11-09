package config

import (
	"testing"

	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/stretchr/testify/assert"
)

func TestURLConfig(t *testing.T) {
	u := NewURLConfig(EnvironmentOf(MapFetcher(map[string][]string{
		"http.key":                                []string{"root", "root-2"},
		"http.https://host.com.key":               []string{"host", "host-2"},
		"http.https://user@host.com/a.key":        []string{"user-a", "user-b"},
		"http.https://user@host.com.key":          []string{"user", "user-2"},
		"http.https://host.com/a.key":             []string{"host-a", "host-b"},
		"http.https://host.com:8080.key":          []string{"port", "port-2"},
		"http.https://host.com/repo.git.key":      []string{".git"},
		"http.https://host.com/repo.key":          []string{"no .git"},
		"http.https://host.com/repo2.key":         []string{"no .git"},
		"http.http://host.com/repo.key":           []string{"http"},
		"http.https://host.com:443/repo3.git.key": []string{"port"},
		"http.ssh://host.com:22/repo3.git.key":    []string{"ssh-port"},
		"http.https://host.*/a.key":               []string{"wild"},
		"httpXhttps://host.*/aXkey":               []string{"invalid"},
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
		"https://host.com:443/repo2.git/info/lfs":     "no .git",
		"https://host.com/repo2.git/info":             "host-2", // doesn't match /.git/info/lfs\Z/
		"https://host.com/repo2.git":                  "host-2", // ditto
		"https://host.com/repo3.git/info/lfs":         "port",
		"ssh://host.com/repo3.git/info/lfs":           "ssh-port",
		"https://host.com/repo2":                      "no .git",
		"http://host.com/repo":                        "http",
		"http://host.com:80/repo":                     "http",
		"https://host.wild/a/b/c":                     "wild",
	}

	for rawurl, expected := range getOne {
		value, _ := u.Get("http", rawurl, "key")
		assert.Equal(t, expected, value, "get one: "+rawurl)
	}

	value, _ := u.Get("http", "https://host.wild/a/b/c", "k")
	assert.Equal(t, value, "")
	value, _ = u.Get("ttp", "https://host.wild/a/b/c", "key")
	assert.Equal(t, value, "")

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

func TestURLConfigEqualMatchesFollowConfigOrder(t *testing.T) {
	for _, test := range []struct {
		name   string
		config string
		want   string
	}{
		{
			name: "trailing slash last",
			config: "http.https://host.com.key=without slash\n" +
				"http.https://host.com/.key=with slash",
			want: "with slash",
		},
		{
			name: "trailing slash first",
			config: "http.https://host.com/.key=with slash\n" +
				"http.https://host.com.key=without slash",
			want: "without slash",
		},
		{
			name: "same key repeated last",
			config: "http.https://host.com.key=first\n" +
				"http.https://host.com/.key=with slash\n" +
				"http.https://host.com.key=last",
			want: "last",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			fetcher, _, _ := readGitConfig(git.ParseConfigLines(test.config, false))
			u := NewURLConfig(EnvironmentOf(fetcher))

			value, ok := u.Get("http", "https://host.com/repo", "key")
			assert.True(t, ok)
			assert.Equal(t, test.want, value)
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := map[string]string{
		"AbCdeF://x.Y":                         "abcdef://x.y/",
		"http://x:000000080":                   "http://x/",
		"https://x:000000443":                  "https://x/",
		"http://x:00000800":                    "http://x:800/",
		"X://W/%7e%41^%3a":                     "x://w/~A%5E%3A",
		"x://%41%62(^):%70+d@foo":              "x://Ab(%5E):p+d@foo/",
		"x://y/./":                             "x://y/",
		"x://y/a/./b/.././../c":                "x://y/c",
		"x://y/a/./b/../.././c/":               "x://y/c/",
		"x://y/%2e/":                           "x://y/",
		"x://y/a/%2e./":                        "x://y/",
		"x://y/a/./?/././..":                   "x://y/a/?/././..",
		"x://q/\xc2\x80":                       "x://q/%C2%80",
		"https://USER@EXAMPLE.COM:0443/a/%7eb": "https://USER@example.com/a/~b",
	}

	for rawurl, expected := range tests {
		t.Run(rawurl, func(t *testing.T) {
			normalized, err := normalizeURL(rawurl, false)
			if assert.NoError(t, err) {
				assert.Equal(t, expected, normalized.String())
			}
		})
	}
}

func TestNormalizeURLRejectsInvalidURLs(t *testing.T) {
	tests := []string{
		"",
		"scheme",
		"0test://acme.co",
		"scheme://",
		"scheme://host:0",
		"scheme://host:65536",
		"scheme://host:invalid",
		"scheme://hos%41/",
		"scheme://host/%fg",
		"scheme://host/a/../..",
		"scheme://*.example.com",
	}

	for _, rawurl := range tests {
		t.Run(rawurl, func(t *testing.T) {
			_, err := normalizeURL(rawurl, false)
			assert.Error(t, err)
		})
	}

	normalized, err := normalizeURL("scheme://*.example.com", true)
	if assert.NoError(t, err) {
		assert.Equal(t, "scheme://*.example.com/", normalized.String())
	}
}

func TestURLConfigMatchesNormalizedURLs(t *testing.T) {
	u := NewURLConfig(EnvironmentOf(MapFetcher(map[string][]string{
		"http.https://example.com.key":                 {"root"},
		"http.HTTPS://EXAMPLE.COM:0443/a/./b.key":      {"canonical path"},
		"http.https://example.com/users/%7euser.key":   {"escaped unreserved"},
		"http.https://example.com/a%2fb.key":           {"escaped slash"},
		"http.https://*.example.com/wildcard/%2e.key":  {"wildcard"},
		"http.https://user%31@example.com/private.key": {"username"},
	})))

	tests := map[string]string{
		"https://example.com/a/b/c":               "canonical path",
		"https://example.com/users/~user/profile": "escaped unreserved",
		"https://example.com/a/c":                 "root",
		"https://sub.example.com/wildcard/child":  "wildcard",
		"https://user1@example.com/private/repo":  "username",
	}

	for rawurl, expected := range tests {
		t.Run(rawurl, func(t *testing.T) {
			value, ok := u.Get("http", rawurl, "key")
			assert.True(t, ok)
			assert.Equal(t, expected, value)
		})
	}
}

func TestGitFetcherSortedAll(t *testing.T) {
	fetcher, _, _ := readGitConfig(git.ParseConfigLines(
		"http.https://host.com.key=first\n"+
			"http.https://host.com/.key=middle\n"+
			"http.https://host.com.key=last",
		false,
	))

	assert.Equal(t, []EnvironmentEntry{
		{Key: "http.https://host.com/.key", Values: []string{"middle"}},
		{Key: "http.https://host.com.key", Values: []string{"first", "last"}},
	}, fetcher.SortedAll())
}
