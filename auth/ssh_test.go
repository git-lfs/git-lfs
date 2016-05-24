package auth

import (
	"path/filepath"
	"testing"

	"github.com/github/git-lfs/config"
	"github.com/technoweenie/assert"
)

func TestSSHGetExeAndArgsSsh(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSH := config.Config.Getenv("GIT_SSH")
	config.Config.Setenv("GIT_SSH", "")
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsSshCustomPort(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSH := config.Config.Getenv("GIT_SSH")
	config.Config.Setenv("GIT_SSH", "")
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"-p", "8888", "user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsPlink(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSH := config.Config.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "plink.exe")
	config.Config.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsPlinkCustomPort(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSH := config.Config.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "plink")
	config.Config.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsTortoisePlink(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSH := config.Config.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink.exe")
	config.Config.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsTortoisePlinkCustomPort(t *testing.T) {
	endpoint := config.Config.Endpoint("download")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSH := config.Config.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink")
	config.Config.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "-P", "8888", "user@foo.com"}, args)

	config.Config.Setenv("GIT_SSH", oldGITSSH)
}
