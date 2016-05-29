package auth

import (
	"path/filepath"
	"testing"

	"github.com/github/git-lfs/config"
	"github.com/stretchr/testify/assert"
)

func TestSSHGetExeAndArgsSsh(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSHCommand := config.Config.Getenv("GIT_SSH_COMMAND")
	config.Config.Setenv("GIT_SSH_COMMAND", "")
	oldGITSSH := config.Config.Getenv("GIT_SSH")
	config.Config.Setenv("GIT_SSH", "")
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH", oldGITSSH)
	config.Config.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsSshCustomPort(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSHCommand := config.Config.Getenv("GIT_SSH_COMMAND")
	config.Config.Setenv("GIT_SSH_COMMAND", "")
	oldGITSSH := config.Config.Getenv("GIT_SSH")
	config.Config.Setenv("GIT_SSH", "")
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"-p", "8888", "user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH", oldGITSSH)
	config.Config.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsPlink(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSHCommand := config.Config.Getenv("GIT_SSH_COMMAND")
	config.Config.Setenv("GIT_SSH_COMMAND", "")
	oldGITSSH := config.Config.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "plink.exe")
	config.Config.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH", oldGITSSH)
	config.Config.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsPlinkCustomPort(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSHCommand := config.Config.Getenv("GIT_SSH_COMMAND")
	config.Config.Setenv("GIT_SSH_COMMAND", "")
	oldGITSSH := config.Config.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "plink")
	config.Config.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH", oldGITSSH)
	config.Config.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsTortoisePlink(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSHCommand := config.Config.Getenv("GIT_SSH_COMMAND")
	config.Config.Setenv("GIT_SSH_COMMAND", "")
	oldGITSSH := config.Config.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink.exe")
	config.Config.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH", oldGITSSH)
	config.Config.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsTortoisePlinkCustomPort(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSHCommand := config.Config.Getenv("GIT_SSH_COMMAND")
	config.Config.Setenv("GIT_SSH_COMMAND", "")
	oldGITSSH := config.Config.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink")
	config.Config.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "-P", "8888", "user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH", oldGITSSH)
	config.Config.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsSshCommandPrecedence(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSHCommand := config.Config.Getenv("GIT_SSH_COMMAND")
	config.Config.Setenv("GIT_SSH_COMMAND", "sshcmd")
	oldGITSSH := config.Config.Getenv("GIT_SSH")
	config.Config.Setenv("GIT_SSH", "bad")
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, "sshcmd", exe)
	assert.Equal(t, []string{"user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH", oldGITSSH)
	config.Config.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsSshCommandArgs(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSHCommand := config.Config.Getenv("GIT_SSH_COMMAND")
	config.Config.Setenv("GIT_SSH_COMMAND", "sshcmd --args 1")
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, "sshcmd", exe)
	assert.Equal(t, []string{"--args", "1", "user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsSshCommandCustomPort(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSHCommand := config.Config.Getenv("GIT_SSH_COMMAND")
	config.Config.Setenv("GIT_SSH_COMMAND", "sshcmd")
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, "sshcmd", exe)
	assert.Equal(t, []string{"-p", "8888", "user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsPlinkCommand(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSHCommand := config.Config.Getenv("GIT_SSH_COMMAND")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "plink.exe")
	config.Config.Setenv("GIT_SSH_COMMAND", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsPlinkCommandCustomPort(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSHCommand := config.Config.Getenv("GIT_SSH_COMMAND")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "plink")
	config.Config.Setenv("GIT_SSH_COMMAND", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsTortoisePlinkCommand(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSHCommand := config.Config.Getenv("GIT_SSH_COMMAND")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink.exe")
	config.Config.Setenv("GIT_SSH_COMMAND", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}

func TestSSHGetExeAndArgsTortoisePlinkCommandCustomPort(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSHCommand := config.Config.Getenv("GIT_SSH_COMMAND")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink")
	config.Config.Setenv("GIT_SSH_COMMAND", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "-P", "8888", "user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH_COMMAND", oldGITSSHCommand)
}
