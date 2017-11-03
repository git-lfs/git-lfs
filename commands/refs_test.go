package commands

import (
	"testing"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/git"
	"github.com/stretchr/testify/assert"
)

func TestRefUpdateDefault(t *testing.T) {
	pushModes := []string{"simple", ""}
	for _, pushMode := range pushModes {
		cfg := config.NewFrom(config.Values{
			Git: map[string][]string{
				"push.default":       []string{pushMode},
				"branch.left.remote": []string{"ignore"},
				"branch.left.merge":  []string{"me"},
			},
		})

		u := newRefUpdate(cfg.Git, "origin", git.ParseRef("left", ""), nil)
		assert.Equal(t, "left", u.Right().Name, "pushmode=%q", pushMode)
	}
}

func TestRefUpdateTrackedDefault(t *testing.T) {
	pushModes := []string{"simple", "upstream", "tracking", ""}
	for _, pushMode := range pushModes {
		cfg := config.NewFrom(config.Values{
			Git: map[string][]string{
				"push.default":       []string{pushMode},
				"branch.left.remote": []string{"origin"},
				"branch.left.merge":  []string{"tracked"},
			},
		})

		u := newRefUpdate(cfg.Git, "origin", git.ParseRef("left", ""), nil)
		assert.Equal(t, "tracked", u.Right().Name, "pushmode=%s", pushMode)
	}
}

func TestRefUpdateCurrentDefault(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Git: map[string][]string{
			"push.default":       []string{"current"},
			"branch.left.remote": []string{"origin"},
			"branch.left.merge":  []string{"tracked"},
		},
	})

	u := newRefUpdate(cfg.Git, "origin", git.ParseRef("left", ""), nil)
	assert.Equal(t, "left", u.Right().Name)
}

func TestRefUpdateExplicitLeftAndRight(t *testing.T) {
	u := newRefUpdate(nil, "", git.ParseRef("left", "abc123"), git.ParseRef("right", "def456"))
	assert.Equal(t, "left", u.Left().Name)
	assert.Equal(t, "abc123", u.Left().Sha)
	assert.Equal(t, "abc123", u.LeftCommitish())
	assert.Equal(t, "right", u.Right().Name)
	assert.Equal(t, "def456", u.Right().Sha)
	assert.Equal(t, "def456", u.RightCommitish())

	u = newRefUpdate(nil, "", git.ParseRef("left", ""), git.ParseRef("right", ""))
	assert.Equal(t, "left", u.Left().Name)
	assert.Equal(t, "", u.Left().Sha)
	assert.Equal(t, "left", u.LeftCommitish())
	assert.Equal(t, "right", u.Right().Name)
	assert.Equal(t, "", u.Right().Sha)
	assert.Equal(t, "right", u.RightCommitish())
}
