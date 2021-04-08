package config

import (
	"os"
	"testing"
	"time"

	"github.com/git-lfs/git-lfs/git"
	"github.com/stretchr/testify/assert"
)

func TestRemoteDefault(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"branch.unused.remote":     []string{"a"},
			"branch.unused.pushremote": []string{"b"},
		},
	})
	assert.Equal(t, "origin", cfg.Remote())
	assert.Equal(t, "origin", cfg.PushRemote())
}

func TestRemoteBranchConfig(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"branch.master.remote":    []string{"a"},
			"branch.other.pushremote": []string{"b"},
		},
	})
	cfg.ref = &git.Ref{Name: "master"}

	assert.Equal(t, "a", cfg.Remote())
	assert.Equal(t, "a", cfg.PushRemote())
}

func TestRemotePushDefault(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"branch.master.remote":    []string{"a"},
			"remote.pushdefault":      []string{"b"},
			"branch.other.pushremote": []string{"c"},
		},
	})
	cfg.ref = &git.Ref{Name: "master"}

	assert.Equal(t, "a", cfg.Remote())
	assert.Equal(t, "b", cfg.PushRemote())
}

func TestRemoteBranchPushDefault(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"branch.master.remote":     []string{"a"},
			"remote.pushdefault":       []string{"b"},
			"branch.master.pushremote": []string{"c"},
		},
	})
	cfg.ref = &git.Ref{Name: "master"}

	assert.Equal(t, "a", cfg.Remote())
	assert.Equal(t, "c", cfg.PushRemote())
}

func TestLFSDefault(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"remote.lfspushdefault": []string{"a"},
			"remote.pushdefault":    []string{"b"},
			"remote.lfsdefault":     []string{"c"},
		},
	})

	assert.Equal(t, "c", cfg.Remote())
	assert.Equal(t, "a", cfg.PushRemote())
}

func TestLFSDefaultSimple(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"remote.lfsdefault": []string{"a"},
		},
	})

	assert.Equal(t, "a", cfg.Remote())
	assert.Equal(t, "a", cfg.PushRemote())
}

func TestLFSDefaultBranch(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"branch.main.remote":     []string{"a"},
			"remote.pushdefault":     []string{"b"},
			"branch.main.pushremote": []string{"c"},
			"remote.lfspushdefault":  []string{"d"},
			"remote.lfsdefault":      []string{"e"},
		},
	})
	cfg.ref = &git.Ref{Name: "main"}

	assert.Equal(t, "a", cfg.Remote())
	assert.Equal(t, "c", cfg.PushRemote())
}

func TestBasicTransfersOnlySetValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"lfs.basictransfersonly": []string{"true"},
		},
	})

	b := cfg.BasicTransfersOnly()
	assert.Equal(t, true, b)
}

func TestBasicTransfersOnlyDefault(t *testing.T) {
	cfg := NewFrom(Values{})

	b := cfg.BasicTransfersOnly()
	assert.Equal(t, false, b)
}

func TestBasicTransfersOnlyInvalidValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"lfs.basictransfersonly": []string{"wat"},
		},
	})

	b := cfg.BasicTransfersOnly()
	assert.Equal(t, false, b)
}

func TestTusTransfersAllowedSetValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"lfs.tustransfers": []string{"true"},
		},
	})

	b := cfg.TusTransfersAllowed()
	assert.Equal(t, true, b)
}

func TestTusTransfersAllowedDefault(t *testing.T) {
	cfg := NewFrom(Values{})

	b := cfg.TusTransfersAllowed()
	assert.Equal(t, false, b)
}

func TestTusTransfersAllowedInvalidValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"lfs.tustransfers": []string{"wat"},
		},
	})

	b := cfg.TusTransfersAllowed()
	assert.Equal(t, false, b)
}

func TestLoadValidExtension(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"lfs.extension.foo.clean":    []string{"foo-clean %f"},
			"lfs.extension.foo.smudge":   []string{"foo-smudge %f"},
			"lfs.extension.foo.priority": []string{"2"},
		},
	})

	ext := cfg.Extensions()["foo"]

	assert.Equal(t, "foo", ext.Name)
	assert.Equal(t, "foo-clean %f", ext.Clean)
	assert.Equal(t, "foo-smudge %f", ext.Smudge)
	assert.Equal(t, 2, ext.Priority)
}

func TestLoadInvalidExtension(t *testing.T) {
	cfg := NewFrom(Values{})
	ext := cfg.Extensions()["foo"]

	assert.Equal(t, "", ext.Name)
	assert.Equal(t, "", ext.Clean)
	assert.Equal(t, "", ext.Smudge)
	assert.Equal(t, 0, ext.Priority)
}

func TestFetchIncludeExcludesAreCleaned(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"lfs.fetchinclude": []string{"/path/to/clean/"},
			"lfs.fetchexclude": []string{"/other/path/to/clean/"},
		},
	})

	assert.Equal(t, []string{"/path/to/clean"}, cfg.FetchIncludePaths())
	assert.Equal(t, []string{"/other/path/to/clean"}, cfg.FetchExcludePaths())
}

func TestRepositoryPermissions(t *testing.T) {
	perms := 0666 & ^umask()

	values := map[string]int{
		"group":     0660,
		"true":      0660,
		"1":         0660,
		"YES":       0660,
		"all":       0664,
		"world":     0664,
		"everybody": 0664,
		"2":         0664,
		"false":     perms,
		"umask":     perms,
		"0":         perms,
		"NO":        perms,
		"this does not remotely look like a valid value": perms,
		"0664": 0664,
		"0666": 0666,
		"0600": 0600,
		"0660": 0660,
		"0644": 0644,
	}

	for key, val := range values {
		cfg := NewFrom(Values{
			Git: map[string][]string{
				"core.sharedrepository": []string{key},
			},
		})
		assert.Equal(t, os.FileMode(val), cfg.RepositoryPermissions(false))
	}
}

func TestRepositoryPermissionsExectable(t *testing.T) {
	perms := 0777 & ^umask()

	values := map[string]int{
		"group":     0770,
		"true":      0770,
		"1":         0770,
		"YES":       0770,
		"all":       0775,
		"world":     0775,
		"everybody": 0775,
		"2":         0775,
		"false":     perms,
		"umask":     perms,
		"0":         perms,
		"NO":        perms,
		"this does not remotely look like a valid value": perms,
		"0664": 0775,
		"0666": 0777,
		"0600": 0700,
		"0660": 0770,
		"0644": 0755,
	}

	for key, val := range values {
		cfg := NewFrom(Values{
			Git: map[string][]string{
				"core.sharedrepository": []string{key},
			},
		})
		assert.Equal(t, os.FileMode(val), cfg.RepositoryPermissions(true))
	}
}

func TestCurrentUser(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"user.name":  []string{"Pat Doe"},
			"user.email": []string{"pdoe@example.org"},
		},
		Os: map[string][]string{
			"EMAIL": []string{"pdoe@example.com"},
		},
	})

	name, email := cfg.CurrentCommitter()
	assert.Equal(t, name, "Pat Doe")
	assert.Equal(t, email, "pdoe@example.org")

	cfg = NewFrom(Values{
		Git: map[string][]string{
			"user.name": []string{"Pat Doe"},
		},
		Os: map[string][]string{
			"EMAIL": []string{"pdoe@example.com"},
		},
	})

	name, email = cfg.CurrentCommitter()
	assert.Equal(t, name, "Pat Doe")
	assert.Equal(t, email, "pdoe@example.com")

	cfg = NewFrom(Values{
		Git: map[string][]string{
			"user.name":  []string{"Pat Doe"},
			"user.email": []string{"pdoe@example.org"},
		},
		Os: map[string][]string{
			"GIT_COMMITTER_NAME":  []string{"Sam Roe"},
			"GIT_COMMITTER_EMAIL": []string{"sroe@example.net"},
			"EMAIL":               []string{"pdoe@example.com"},
		},
	})

	name, email = cfg.CurrentCommitter()
	assert.Equal(t, name, "Sam Roe")
	assert.Equal(t, email, "sroe@example.net")

	cfg = NewFrom(Values{
		Git: map[string][]string{
			"user.name":  []string{"Pat Doe"},
			"user.email": []string{"pdoe@example.org"},
		},
		Os: map[string][]string{
			"GIT_AUTHOR_NAME":  []string{"Sam Roe"},
			"GIT_AUTHOR_EMAIL": []string{"sroe@example.net"},
			"EMAIL":            []string{"pdoe@example.com"},
		},
	})

	name, email = cfg.CurrentCommitter()
	assert.Equal(t, name, "Pat Doe")
	assert.Equal(t, email, "pdoe@example.org")

	name, email = cfg.CurrentAuthor()
	assert.Equal(t, name, "Sam Roe")
	assert.Equal(t, email, "sroe@example.net")
}

func TestCurrentTimestamp(t *testing.T) {
	m := map[string]string{
		"1136239445 -0700":                "2006-01-02T15:04:05-07:00",
		"Mon, 02 Jan 2006 15:04:05 -0700": "2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05-0700":        "2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05-0700":        "2006-01-02T15:04:05-07:00",
		"2006.01.02T15:04:05-0700":        "2006-01-02T15:04:05-07:00",
		"2006.01.02 15:04:05-0700":        "2006-01-02T15:04:05-07:00",
		"01/02/2006T15:04:05-0700":        "2006-01-02T15:04:05-07:00",
		"01/02/2006 15:04:05-0700":        "2006-01-02T15:04:05-07:00",
		"02.01.2006T15:04:05-0700":        "2006-01-02T15:04:05-07:00",
		"02.01.2006 15:04:05-0700":        "2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05Z":            "2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05Z":            "2006-01-02T15:04:05Z",
		"2006.01.02T15:04:05Z":            "2006-01-02T15:04:05Z",
		"2006.01.02 15:04:05Z":            "2006-01-02T15:04:05Z",
		"01/02/2006T15:04:05Z":            "2006-01-02T15:04:05Z",
		"01/02/2006 15:04:05Z":            "2006-01-02T15:04:05Z",
		"02.01.2006T15:04:05Z":            "2006-01-02T15:04:05Z",
		"02.01.2006 15:04:05Z":            "2006-01-02T15:04:05Z",
		"not a date":                      "default",
		"":                                "default",
	}

	for val, res := range m {
		cfg := NewFrom(Values{
			Os: map[string][]string{
				"GIT_COMMITTER_DATE": []string{val},
			},
		})
		date := cfg.CurrentCommitterTimestamp()

		if res == "default" {
			assert.Equal(t, date, cfg.timestamp)
		} else {
			assert.Equal(t, date.Format(time.RFC3339), res)
		}
	}
}

func TestRemoteNameWithDotDefault(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"remote.name.with.dot.url": []string{"http://remote.url/repo"},
		},
	})

	assert.Equal(t, "name.with.dot", cfg.Remote())
}
