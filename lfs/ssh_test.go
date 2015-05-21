package lfs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmizerany/assert"
)

func TestSSHGetExeAndArgsSsh(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSH := os.Getenv("GIT_SSH")
	os.Setenv("GIT_SSH", "")
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsSshCustomPort(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSH := os.Getenv("GIT_SSH")
	os.Setenv("GIT_SSH", "")
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"-p", "8888", "user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsPlink(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSH := os.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "plink.exe")
	os.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsPlinkCustomPort(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSH := os.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "plink")
	os.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsTortoisePlink(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSH := os.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink.exe")
	os.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsTortoisePlinkCustomPort(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSH := os.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink")
	os.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "-P", "8888", "user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}
