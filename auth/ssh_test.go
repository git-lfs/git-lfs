package auth

import (
	"path/filepath"
	"testing"

	"github.com/github/git-lfs/config"
	"github.com/stretchr/testify/assert"
)

func TestSSHGetExeAndArgsSsh(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"GIT_SSH_COMMAND": "",
			"GIT_SSH":         "",
		},
	})

	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"

	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCustomPort(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"GIT_SSH_COMMAND": "",
			"GIT_SSH":         "",
		},
	})

	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"

	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"-p", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlink(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "plink.exe")

	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"GIT_SSH_COMMAND": "",
			"GIT_SSH":         plink,
		},
	})

	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"

	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlinkCustomPort(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "plink")

	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"GIT_SSH_COMMAND": "",
			"GIT_SSH":         plink,
		},
	})

	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"

	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlink(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink.exe")

	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"GIT_SSH_COMMAND": "",
			"GIT_SSH":         plink,
		},
	})

	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"

	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlinkCustomPort(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink")

	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"GIT_SSH_COMMAND": "",
			"GIT_SSH":         plink,
		},
	})

	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"

	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "-P", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCommandPrecedence(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"GIT_SSH_COMMAND": "sshcmd",
			"GIT_SSH":         "bad",
		},
	})

	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"

	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, "sshcmd", exe)
	assert.Equal(t, []string{"user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCommandArgs(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"GIT_SSH_COMMAND": "sshcmd --args 1",
		},
	})

	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"

	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, "sshcmd", exe)
	assert.Equal(t, []string{"--args", "1", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCommandCustomPort(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"GIT_SSH_COMMAND": "sshcmd",
		},
	})

	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"

	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, "sshcmd", exe)
	assert.Equal(t, []string{"-p", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlinkCommand(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "plink.exe")

	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"GIT_SSH_COMMAND": plink,
		},
	})

	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"

	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlinkCommandCustomPort(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "plink")

	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"GIT_SSH_COMMAND": plink,
		},
	})

	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"

	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlinkCommand(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink.exe")

	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"GIT_SSH_COMMAND": plink,
		},
	})

	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"

	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlinkCommandCustomPort(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink")

	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"GIT_SSH_COMMAND": plink,
		},
	})

	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"

	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "-P", "8888", "user@foo.com"}, args)
}
