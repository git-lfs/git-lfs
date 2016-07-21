package auth

import (
	"path/filepath"
	"testing"

	"github.com/github/git-lfs/config"
	"github.com/stretchr/testify/assert"
)

func TestSSHGetExeAndArgsSsh(t *testing.T) {
	cfg := config.NewConfig()
	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSHCommand := cfg.Getenv("GIT_SSH_COMMAND")
	cfg.Setenv("GIT_SSH_COMMAND", "")
	oldGITSSH := cfg.Getenv("GIT_SSH")
	cfg.Setenv("GIT_SSH", "")
	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"user@foo.com"}, args)

	cfg.Setenv("GIT_SSH", oldGITSSH)
	cfg.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsSshCustomPort(t *testing.T) {
	cfg := config.NewConfig()
	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSHCommand := cfg.Getenv("GIT_SSH_COMMAND")
	cfg.Setenv("GIT_SSH_COMMAND", "")
	oldGITSSH := cfg.Getenv("GIT_SSH")
	cfg.Setenv("GIT_SSH", "")
	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"-p", "8888", "user@foo.com"}, args)

	cfg.Setenv("GIT_SSH", oldGITSSH)
	cfg.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsPlink(t *testing.T) {
	cfg := config.NewConfig()
	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSHCommand := cfg.Getenv("GIT_SSH_COMMAND")
	cfg.Setenv("GIT_SSH_COMMAND", "")
	oldGITSSH := cfg.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "plink.exe")
	cfg.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"user@foo.com"}, args)

	cfg.Setenv("GIT_SSH", oldGITSSH)
	cfg.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsPlinkCustomPort(t *testing.T) {
	cfg := config.NewConfig()
	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSHCommand := cfg.Getenv("GIT_SSH_COMMAND")
	cfg.Setenv("GIT_SSH_COMMAND", "")
	oldGITSSH := cfg.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "plink")
	cfg.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)

	cfg.Setenv("GIT_SSH", oldGITSSH)
	cfg.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsTortoisePlink(t *testing.T) {
	cfg := config.NewConfig()
	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSHCommand := cfg.Getenv("GIT_SSH_COMMAND")
	cfg.Setenv("GIT_SSH_COMMAND", "")
	oldGITSSH := cfg.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink.exe")
	cfg.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "user@foo.com"}, args)

	cfg.Setenv("GIT_SSH", oldGITSSH)
	cfg.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsTortoisePlinkCustomPort(t *testing.T) {
	cfg := config.NewConfig()
	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSHCommand := cfg.Getenv("GIT_SSH_COMMAND")
	cfg.Setenv("GIT_SSH_COMMAND", "")
	oldGITSSH := cfg.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink")
	cfg.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "-P", "8888", "user@foo.com"}, args)

	cfg.Setenv("GIT_SSH", oldGITSSH)
	cfg.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsSshCommandPrecedence(t *testing.T) {
	cfg := config.NewConfig()
	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSHCommand := cfg.Getenv("GIT_SSH_COMMAND")
	cfg.Setenv("GIT_SSH_COMMAND", "sshcmd")
	oldGITSSH := cfg.Getenv("GIT_SSH")
	cfg.Setenv("GIT_SSH", "bad")
	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, "sshcmd", exe)
	assert.Equal(t, []string{"user@foo.com"}, args)

	cfg.Setenv("GIT_SSH", oldGITSSH)
	cfg.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsSshCommandArgs(t *testing.T) {
	cfg := config.NewConfig()
	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSHCommand := cfg.Getenv("GIT_SSH_COMMAND")
	cfg.Setenv("GIT_SSH_COMMAND", "sshcmd --args 1")
	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, "sshcmd", exe)
	assert.Equal(t, []string{"--args", "1", "user@foo.com"}, args)

	cfg.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsSshCommandCustomPort(t *testing.T) {
	cfg := config.NewConfig()
	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSHCommand := cfg.Getenv("GIT_SSH_COMMAND")
	cfg.Setenv("GIT_SSH_COMMAND", "sshcmd")
	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, "sshcmd", exe)
	assert.Equal(t, []string{"-p", "8888", "user@foo.com"}, args)

	cfg.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsPlinkCommand(t *testing.T) {
	cfg := config.NewConfig()
	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSHCommand := cfg.Getenv("GIT_SSH_COMMAND")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "plink.exe")
	cfg.Setenv("GIT_SSH_COMMAND", plink)
	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"user@foo.com"}, args)

	cfg.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsPlinkCommandCustomPort(t *testing.T) {
	cfg := config.NewConfig()
	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSHCommand := cfg.Getenv("GIT_SSH_COMMAND")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "plink")
	cfg.Setenv("GIT_SSH_COMMAND", plink)
	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)

	cfg.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsTortoisePlinkCommand(t *testing.T) {
	cfg := config.NewConfig()
	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSHCommand := cfg.Getenv("GIT_SSH_COMMAND")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink.exe")
	cfg.Setenv("GIT_SSH_COMMAND", plink)
	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "user@foo.com"}, args)

	cfg.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsTortoisePlinkCommandCustomPort(t *testing.T) {
	cfg := config.NewConfig()
	endpoint := cfg.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSHCommand := cfg.Getenv("GIT_SSH_COMMAND")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink")
	cfg.Setenv("GIT_SSH_COMMAND", plink)
	exe, args := sshGetExeAndArgs(cfg, endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "-P", "8888", "user@foo.com"}, args)

	cfg.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}
