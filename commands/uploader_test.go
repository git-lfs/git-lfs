package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type LockingSupportTestCase struct {
	Given           string
	ExpectedToMatch bool
}

func (l *LockingSupportTestCase) Assert(t *testing.T) {
	assert.Equal(t, l.ExpectedToMatch, supportsLockingAPI(l.Given))
}

func TestSupportedLockingHosts(t *testing.T) {
	for desc, c := range map[string]*LockingSupportTestCase{
		"https with path prefix":        {"https://github.com/ttaylorr/dotfiles.git/info/lfs", true},
		"https with root":               {"https://github.com/ttaylorr/dotfiles", true},
		"http with path prefix":         {"http://github.com/ttaylorr/dotfiles.git/info/lfs", false},
		"http with root":                {"http://github.com/ttaylorr/dotfiles", false},
		"ssh with path prefix":          {"ssh://github.com/ttaylorr/dotfiles.git/info/lfs", true},
		"ssh with root":                 {"ssh://github.com/ttaylorr/dotfiles", true},
		"ssh with user and path prefix": {"ssh://git@github.com/ttaylorr/dotfiles.git/info/lfs", true},
		"ssh with user and root":        {"ssh://git@github.com/ttaylorr/dotfiles", true},
	} {
		t.Run(desc, c.Assert)
	}
}
