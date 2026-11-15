package ssh

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A control directory created for a multiplexed connection must be removed when
// the transfer shuts down; otherwise every pure SSH transfer leaves an empty
// sock-* directory behind in the runtime directory.
func TestShutdownRemovesCreatedControlDir(t *testing.T) {
	runtimeDir := t.TempDir()
	env := config.EnvironmentOf(config.MapFetcher(map[string][]string{
		"XDG_RUNTIME_DIR": {runtimeDir},
	}))

	controlDir, err := getControlDir(env)
	require.NoError(t, err)
	require.DirExists(t, controlDir)
	require.Equal(t, runtimeDir, filepath.Dir(controlDir))

	transfer := &SSHTransfer{
		lock:       &sync.RWMutex{},
		osEnv:      env,
		gitEnv:     config.EnvironmentOf(config.MapFetcher(nil)),
		controlDir: controlDir,
	}
	require.NoError(t, transfer.Shutdown())
	require.NoDirExists(t, controlDir)
}

// A control path supplied by the caller belongs to the caller, so shutting a
// transfer down must leave its directory alone.
func TestShutdownKeepsCallerOwnedControlPath(t *testing.T) {
	callerDir := t.TempDir()

	transfer := &SSHTransfer{
		lock:   &sync.RWMutex{},
		osEnv:  config.EnvironmentOf(config.MapFetcher(nil)),
		gitEnv: config.EnvironmentOf(config.MapFetcher(nil)),
	}
	require.NoError(t, transfer.Shutdown())
	require.DirExists(t, callerDir)
}

func TestCreatedControlDir(t *testing.T) {
	sockDir := filepath.Join("run", "sock-123")
	sockPath := filepath.Join(sockDir, "lfs.sock")
	for _, test := range []struct {
		explicitPath string
		multiplexing bool
		controlPath  string
		want         string
	}{
		{"", true, sockPath, sockDir},
		{"", false, "", ""},
		{"", true, "", ""},
		{sockPath, true, sockPath, ""},
	} {
		got := createdControlDir(test.explicitPath, test.multiplexing, test.controlPath)
		assert.Equal(t, test.want, got, "explicit=%q multiplexing=%v controlPath=%q",
			test.explicitPath, test.multiplexing, test.controlPath)
	}
}
