package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRefUpdateDefault(t *testing.T) {
	pushModes := []string{"simple", ""}
	for _, pushMode := range pushModes {
		env := newEnv(map[string][]string{
			"push.default":       []string{pushMode},
			"branch.left.remote": []string{"ignore"},
			"branch.left.merge":  []string{"me"},
		})

		u := NewRefUpdate(env, "origin", ParseRef("refs/heads/left", ""), nil)
		assert.Equal(t, "left", u.Right().Name, "pushmode=%q", pushMode)
		assert.Equal(t, RefTypeLocalBranch, u.Right().Type, "pushmode=%q", pushMode)
	}
}

func TestRefUpdateTrackedDefault(t *testing.T) {
	pushModes := []string{"simple", "upstream", "tracking", ""}
	for _, pushMode := range pushModes {
		env := newEnv(map[string][]string{
			"push.default":       []string{pushMode},
			"branch.left.remote": []string{"origin"},
			"branch.left.merge":  []string{"refs/heads/tracked"},
		})

		u := NewRefUpdate(env, "origin", ParseRef("refs/heads/left", ""), nil)
		assert.Equal(t, "tracked", u.Right().Name, "pushmode=%s", pushMode)
		assert.Equal(t, RefTypeLocalBranch, u.Right().Type, "pushmode=%q", pushMode)
	}
}

func TestRefUpdateCurrentDefault(t *testing.T) {
	env := newEnv(map[string][]string{
		"push.default":       []string{"current"},
		"branch.left.remote": []string{"origin"},
		"branch.left.merge":  []string{"tracked"},
	})

	u := NewRefUpdate(env, "origin", ParseRef("refs/heads/left", ""), nil)
	assert.Equal(t, "left", u.Right().Name)
	assert.Equal(t, RefTypeLocalBranch, u.Right().Type)
}

func TestRefUpdateExplicitLeftAndRight(t *testing.T) {
	u := NewRefUpdate(nil, "", ParseRef("refs/heads/left", "abc123"), ParseRef("refs/heads/right", "def456"))
	assert.Equal(t, "left", u.Left().Name)
	assert.Equal(t, "abc123", u.Left().Sha)
	assert.Equal(t, "abc123", u.LeftCommitish())
	assert.Equal(t, "right", u.Right().Name)
	assert.Equal(t, "def456", u.Right().Sha)
	assert.Equal(t, "def456", u.RightCommitish())

	u = NewRefUpdate(nil, "", ParseRef("refs/heads/left", ""), ParseRef("refs/heads/right", ""))
	assert.Equal(t, "left", u.Left().Name)
	assert.Equal(t, "", u.Left().Sha)
	assert.Equal(t, "left", u.LeftCommitish())
	assert.Equal(t, "right", u.Right().Name)
	assert.Equal(t, "", u.Right().Sha)
	assert.Equal(t, "right", u.RightCommitish())
}

func newEnv(m map[string][]string) *mapEnv {
	return &mapEnv{data: m}
}

type mapEnv struct {
	data map[string][]string
}

func (m *mapEnv) Get(key string) (string, bool) {
	vals, ok := m.data[key]
	if ok && len(vals) > 0 {
		return vals[0], true
	}
	return "", false
}
