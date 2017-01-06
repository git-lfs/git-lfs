package lfsapi

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHGetExeAndArgsSsh(t *testing.T) {
	cli, err := NewClient(TestEnv(map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         "",
	}), nil)
	require.Nil(t, err)

	endpoint := cli.Endpoints.Endpoint("download", "")
	endpoint.SshUserAndHost = "user@foo.com"

	exe, args := sshGetExeAndArgs(cli.OSEnv(), endpoint)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCustomPort(t *testing.T) {
	cli, err := NewClient(TestEnv(map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         "",
	}), nil)
	require.Nil(t, err)

	endpoint := cli.Endpoints.Endpoint("download", "")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"

	exe, args := sshGetExeAndArgs(cli.OSEnv(), endpoint)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"-p", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlink(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "plink.exe")

	cli, err := NewClient(TestEnv(map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
	}), nil)
	require.Nil(t, err)

	endpoint := cli.Endpoints.Endpoint("download", "")
	endpoint.SshUserAndHost = "user@foo.com"

	exe, args := sshGetExeAndArgs(cli.OSEnv(), endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlinkCustomPort(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "plink")

	cli, err := NewClient(TestEnv(map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
	}), nil)
	require.Nil(t, err)

	endpoint := cli.Endpoints.Endpoint("download", "")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"

	exe, args := sshGetExeAndArgs(cli.OSEnv(), endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlink(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink.exe")

	cli, err := NewClient(TestEnv(map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
	}), nil)
	require.Nil(t, err)

	endpoint := cli.Endpoints.Endpoint("download", "")
	endpoint.SshUserAndHost = "user@foo.com"

	exe, args := sshGetExeAndArgs(cli.OSEnv(), endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlinkCustomPort(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink")

	cli, err := NewClient(TestEnv(map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
	}), nil)
	require.Nil(t, err)

	endpoint := cli.Endpoints.Endpoint("download", "")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"

	exe, args := sshGetExeAndArgs(cli.OSEnv(), endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "-P", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCommandPrecedence(t *testing.T) {
	cli, err := NewClient(TestEnv(map[string]string{
		"GIT_SSH_COMMAND": "sshcmd",
		"GIT_SSH":         "bad",
	}), nil)
	require.Nil(t, err)

	endpoint := cli.Endpoints.Endpoint("download", "")
	endpoint.SshUserAndHost = "user@foo.com"

	exe, args := sshGetExeAndArgs(cli.OSEnv(), endpoint)
	assert.Equal(t, "sshcmd", exe)
	assert.Equal(t, []string{"user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCommandArgs(t *testing.T) {
	cli, err := NewClient(TestEnv(map[string]string{
		"GIT_SSH_COMMAND": "sshcmd --args 1",
	}), nil)
	require.Nil(t, err)

	endpoint := cli.Endpoints.Endpoint("download", "")
	endpoint.SshUserAndHost = "user@foo.com"

	exe, args := sshGetExeAndArgs(cli.OSEnv(), endpoint)
	assert.Equal(t, "sshcmd", exe)
	assert.Equal(t, []string{"--args", "1", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCommandCustomPort(t *testing.T) {
	cli, err := NewClient(TestEnv(map[string]string{
		"GIT_SSH_COMMAND": "sshcmd",
	}), nil)
	require.Nil(t, err)

	endpoint := cli.Endpoints.Endpoint("download", "")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"

	exe, args := sshGetExeAndArgs(cli.OSEnv(), endpoint)
	assert.Equal(t, "sshcmd", exe)
	assert.Equal(t, []string{"-p", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlinkCommand(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "plink.exe")

	cli, err := NewClient(TestEnv(map[string]string{
		"GIT_SSH_COMMAND": plink,
	}), nil)
	require.Nil(t, err)

	endpoint := cli.Endpoints.Endpoint("download", "")
	endpoint.SshUserAndHost = "user@foo.com"

	exe, args := sshGetExeAndArgs(cli.OSEnv(), endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlinkCommandCustomPort(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "plink")

	cli, err := NewClient(TestEnv(map[string]string{
		"GIT_SSH_COMMAND": plink,
	}), nil)
	require.Nil(t, err)

	endpoint := cli.Endpoints.Endpoint("download", "")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"

	exe, args := sshGetExeAndArgs(cli.OSEnv(), endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlinkCommand(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink.exe")

	cli, err := NewClient(TestEnv(map[string]string{
		"GIT_SSH_COMMAND": plink,
	}), nil)
	require.Nil(t, err)

	endpoint := cli.Endpoints.Endpoint("download", "")
	endpoint.SshUserAndHost = "user@foo.com"

	exe, args := sshGetExeAndArgs(cli.OSEnv(), endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlinkCommandCustomPort(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink")

	cli, err := NewClient(TestEnv(map[string]string{
		"GIT_SSH_COMMAND": plink,
	}), nil)
	require.Nil(t, err)

	endpoint := cli.Endpoints.Endpoint("download", "")
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"

	exe, args := sshGetExeAndArgs(cli.OSEnv(), endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "-P", "8888", "user@foo.com"}, args)
}
