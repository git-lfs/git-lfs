package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRefUpdateDefault(t *testing.T) {
	pushModes := []string{"simple", ""}
	for _, pushMode := range pushModes {
		env := newEnv(map[string][]string{
			"push.default":        []string{pushMode},
			"branch.local.remote": []string{"ignore"},
			"branch.local.merge":  []string{"me"},
		})

		u := NewRefUpdate(env, "origin", ParseRef("refs/heads/local", ""), nil)
		assert.Equal(t, "local", u.RemoteRef().Name, "pushmode=%q", pushMode)
		assert.Equal(t, RefTypeLocalBranch, u.RemoteRef().Type, "pushmode=%q", pushMode)
	}
}

func TestRefUpdateTrackedDefault(t *testing.T) {
	pushModes := []string{"simple", "upstream", "tracking", ""}
	for _, pushMode := range pushModes {
		env := newEnv(map[string][]string{
			"push.default":        []string{pushMode},
			"branch.local.remote": []string{"origin"},
			"branch.local.merge":  []string{"refs/heads/tracked"},
		})

		u := NewRefUpdate(env, "origin", ParseRef("refs/heads/local", ""), nil)
		assert.Equal(t, "tracked", u.RemoteRef().Name, "pushmode=%s", pushMode)
		assert.Equal(t, RefTypeLocalBranch, u.RemoteRef().Type, "pushmode=%q", pushMode)
	}
}

func TestRefUpdateCurrentDefault(t *testing.T) {
	env := newEnv(map[string][]string{
		"push.default":        []string{"current"},
		"branch.local.remote": []string{"origin"},
		"branch.local.merge":  []string{"tracked"},
	})

	u := NewRefUpdate(env, "origin", ParseRef("refs/heads/local", ""), nil)
	assert.Equal(t, "local", u.RemoteRef().Name)
	assert.Equal(t, RefTypeLocalBranch, u.RemoteRef().Type)
}

func TestRefUpdateExplicitLocalAndRemoteRefs(t *testing.T) {
	u := NewRefUpdate(nil, "", ParseRef("refs/heads/local", "abc123"), ParseRef("refs/heads/remote", "def456"))
	assert.Equal(t, "local", u.LocalRef().Name)
	assert.Equal(t, "abc123", u.LocalRef().Sha)
	assert.Equal(t, "abc123", u.LocalRefCommitish())
	assert.Equal(t, "remote", u.RemoteRef().Name)
	assert.Equal(t, "def456", u.RemoteRef().Sha)
	assert.Equal(t, "def456", u.RemoteRefCommitish())

	u = NewRefUpdate(nil, "", ParseRef("refs/heads/local", ""), ParseRef("refs/heads/remote", ""))
	assert.Equal(t, "local", u.LocalRef().Name)
	assert.Equal(t, "", u.LocalRef().Sha)
	assert.Equal(t, "local", u.LocalRefCommitish())
	assert.Equal(t, "remote", u.RemoteRef().Name)
	assert.Equal(t, "", u.RemoteRef().Sha)
	assert.Equal(t, "remote", u.RemoteRefCommitish())
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
