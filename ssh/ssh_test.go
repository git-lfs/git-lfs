package ssh_test

import (
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/git-lfs/git-lfs/v3/lfshttp"
	"github.com/git-lfs/git-lfs/v3/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHGetLFSExeAndArgs(t *testing.T) {
	cli, err := lfshttp.NewClient(nil)
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Path = "user/repo"

	exe, args, _, _ := ssh.GetLFSExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, "git-lfs-authenticate", "download", false, "")
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{
		"user@foo.com",
		"git-lfs-authenticate user/repo download",
	}, args)

	exe, args, _, _ = ssh.GetLFSExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, "git-lfs-authenticate", "upload", false, "")
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{
		"user@foo.com",
		"git-lfs-authenticate user/repo upload",
	}, args)
}

func TestSSHGetExeAndArgsSsh(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         "",
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCustomPort(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         "",
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"-p", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshNoMultiplexing(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         "",
	}, map[string]string{
		"lfs.ssh.automultiplex": "false",
	}))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, baseargs, needShell, multiplexing, controlPath := ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, true, "")
	exe, args := ssh.FormatArgs(exe, baseargs, needShell, multiplexing, controlPath)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, false, multiplexing)
	assert.Equal(t, []string{"user@foo.com"}, args)
	assert.Empty(t, controlPath)
}

func TestSSHGetExeAndArgsSshMultiplexingMaster(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         "",
	}, map[string]string{
		"lfs.ssh.automultiplex": "true",
	}))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, baseargs, needShell, multiplexing, controlPath := ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, true, "")
	exe, args := ssh.FormatArgs(exe, baseargs, needShell, multiplexing, controlPath)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, true, multiplexing)
	assert.Equal(t, 3, len(args))
	assert.Equal(t, "-oControlMaster=yes", args[0])
	assert.True(t, strings.HasPrefix(args[1], "-oControlPath="))
	assert.Equal(t, "user@foo.com", args[2])
	assert.NotEmpty(t, controlPath)
}

func TestSSHGetExeAndArgsSshMultiplexingExtra(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         "",
	}, map[string]string{
		"lfs.ssh.automultiplex": "true",
	}))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, baseargs, needShell, multiplexing, controlPath := ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, true, "/tmp/lfs/lfs.sock")
	exe, args := ssh.FormatArgs(exe, baseargs, needShell, multiplexing, controlPath)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, true, multiplexing)
	assert.Equal(t, []string{"-oControlMaster=no", "-oControlPath=/tmp/lfs/lfs.sock", "user@foo.com"}, args)
	assert.Equal(t, "/tmp/lfs/lfs.sock", controlPath)
}

func TestSSHGetExeAndArgsPlink(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "plink.exe")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlinkCustomPort(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "plink")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlinkCustomPortExplicitEnvironment(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "ssh")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
		"GIT_SSH_VARIANT": "plink",
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlinkCustomPortExplicitEnvironmentPutty(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "ssh")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
		"GIT_SSH_VARIANT": "putty",
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlinkCustomPortExplicitEnvironmentSsh(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "ssh")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
		"GIT_SSH_VARIANT": "ssh",
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-p", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlink(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink.exe")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlinkCustomPort(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "-P", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlinkCustomPortExplicitEnvironment(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "ssh")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
		"GIT_SSH_VARIANT": "tortoiseplink",
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "-P", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlinkCustomPortExplicitConfig(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "ssh")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
		"GIT_SSH_VARIANT": "tortoiseplink",
	}, map[string]string{
		"ssh.variant": "tortoiseplink",
	}))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "-P", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlinkCustomPortExplicitConfigOverride(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "ssh")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "",
		"GIT_SSH":         plink,
	}, map[string]string{
		"ssh.variant": "putty",
	}))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCommandPrecedence(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "sshcmd",
		"GIT_SSH":         "bad",
		"GIT_SSH_VARIANT": "simple",
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", "sshcmd user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCommandArgs(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "sshcmd --args 1",
		"GIT_SSH_VARIANT": "simple",
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", "sshcmd --args 1 user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCommandArgsWithMixedQuotes(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "sshcmd foo 'bar \"baz\"'",
		"GIT_SSH_VARIANT": "simple",
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", "sshcmd foo 'bar \"baz\"' user@foo.com"}, args)
}

func TestSSHGetExeAndArgsSshCommandCustomPort(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "sshcmd",
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", "sshcmd -p 8888 user@foo.com"}, args)
}

func TestSSHGetExeAndArgsCoreSshCommand(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": "sshcmd --args 2",
	}, map[string]string{
		"core.sshcommand": "sshcmd --args 1",
	}))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", "sshcmd --args 2 user@foo.com"}, args)
}

func TestSSHGetExeAndArgsCoreSshCommandArgsWithMixedQuotes(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"core.sshcommand": "sshcmd foo 'bar \"baz\"'",
	}))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", "sshcmd foo 'bar \"baz\"' user@foo.com"}, args)
}

func TestSSHGetExeAndArgsConfigVersusEnv(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"core.sshcommand": "sshcmd --args 1",
	}))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", "sshcmd --args 1 user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlinkCommand(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "plink.exe")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": plink,
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", plink + " user@foo.com"}, args)
}

func TestSSHGetExeAndArgsPlinkCommandCustomPort(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "plink")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": plink,
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", plink + " -P 8888 user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlinkCommand(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink.exe")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": plink,
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", plink + " -batch user@foo.com"}, args)
}

func TestSSHGetExeAndArgsTortoisePlinkCommandCustomPort(t *testing.T) {
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink")

	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH_COMMAND": plink,
	}, nil))
	require.Nil(t, err)

	meta := ssh.SSHMetadata{}
	meta.UserAndHost = "user@foo.com"
	meta.Port = "8888"

	exe, args := ssh.FormatArgs(ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &meta, false, ""))
	assert.Equal(t, "sh", exe)
	assert.Equal(t, []string{"-c", plink + " -batch -P 8888 user@foo.com"}, args)
}

func TestSSHGetLFSExeAndArgsWithCustomSSH(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH":         "not-ssh",
		"GIT_SSH_VARIANT": "simple",
	}, nil))
	require.Nil(t, err)

	u, err := url.Parse("ssh://git@host.com:12345/repo")
	require.Nil(t, err)

	e := lfshttp.EndpointFromSshUrl(u)
	t.Logf("ENDPOINT: %+v", e)
	assert.Equal(t, "12345", e.SSHMetadata.Port)
	assert.Equal(t, "git@host.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "/repo", e.SSHMetadata.Path)

	exe, args, _, _ := ssh.GetLFSExeAndArgs(cli.OSEnv(), cli.GitEnv(), &e.SSHMetadata, "git-lfs-authenticate", "download", false, "")
	assert.Equal(t, "not-ssh", exe)
	assert.Equal(t, []string{"-p", "12345", "git@host.com", "git-lfs-authenticate /repo download"}, args)
}

func TestSSHGetLFSExeAndArgsInvalidOptionsAsHost(t *testing.T) {
	cli, err := lfshttp.NewClient(nil)
	require.Nil(t, err)

	u, err := url.Parse("ssh://-oProxyCommand=gnome-calculator/repo")
	require.Nil(t, err)
	assert.Equal(t, "-oProxyCommand=gnome-calculator", u.Host)

	e := lfshttp.EndpointFromSshUrl(u)
	t.Logf("ENDPOINT: %+v", e)
	assert.Equal(t, "-oProxyCommand=gnome-calculator", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "/repo", e.SSHMetadata.Path)

	exe, args, _, _ := ssh.GetLFSExeAndArgs(cli.OSEnv(), cli.GitEnv(), &e.SSHMetadata, "git-lfs-authenticate", "download", false, "")
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"--", "-oProxyCommand=gnome-calculator", "git-lfs-authenticate /repo download"}, args)
}

func TestSSHGetLFSExeAndArgsInvalidOptionsAsHostWithCustomSSH(t *testing.T) {
	cli, err := lfshttp.NewClient(lfshttp.NewContext(nil, map[string]string{
		"GIT_SSH":         "not-ssh",
		"GIT_SSH_VARIANT": "simple",
	}, nil))
	require.Nil(t, err)

	u, err := url.Parse("ssh://--oProxyCommand=gnome-calculator/repo")
	require.Nil(t, err)
	assert.Equal(t, "--oProxyCommand=gnome-calculator", u.Host)

	e := lfshttp.EndpointFromSshUrl(u)
	t.Logf("ENDPOINT: %+v", e)
	assert.Equal(t, "--oProxyCommand=gnome-calculator", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "/repo", e.SSHMetadata.Path)

	exe, args, _, _ := ssh.GetLFSExeAndArgs(cli.OSEnv(), cli.GitEnv(), &e.SSHMetadata, "git-lfs-authenticate", "download", false, "")
	assert.Equal(t, "not-ssh", exe)
	assert.Equal(t, []string{"oProxyCommand=gnome-calculator", "git-lfs-authenticate /repo download"}, args)
}

func TestSSHGetExeAndArgsInvalidOptionsAsHost(t *testing.T) {
	cli, err := lfshttp.NewClient(nil)
	require.Nil(t, err)

	u, err := url.Parse("ssh://-oProxyCommand=gnome-calculator")
	require.Nil(t, err)
	assert.Equal(t, "-oProxyCommand=gnome-calculator", u.Host)

	e := lfshttp.EndpointFromSshUrl(u)
	t.Logf("ENDPOINT: %+v", e)
	assert.Equal(t, "-oProxyCommand=gnome-calculator", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)

	exe, args, needShell, _, _ := ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &e.SSHMetadata, false, "")
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"--", "-oProxyCommand=gnome-calculator"}, args)
	assert.Equal(t, false, needShell)
}

func TestSSHGetExeAndArgsInvalidOptionsAsPath(t *testing.T) {
	cli, err := lfshttp.NewClient(nil)
	require.Nil(t, err)

	u, err := url.Parse("ssh://git@git-host.com/-oProxyCommand=gnome-calculator")
	require.Nil(t, err)
	assert.Equal(t, "git-host.com", u.Host)

	e := lfshttp.EndpointFromSshUrl(u)
	t.Logf("ENDPOINT: %+v", e)
	assert.Equal(t, "git@git-host.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "/-oProxyCommand=gnome-calculator", e.SSHMetadata.Path)

	exe, args, needShell, _, _ := ssh.GetExeAndArgs(cli.OSEnv(), cli.GitEnv(), &e.SSHMetadata, false, "")
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"git@git-host.com"}, args)
	assert.Equal(t, false, needShell)
}

func TestParseBareSSHUrl(t *testing.T) {
	e := lfshttp.EndpointFromBareSshUrl("git@git-host.com:repo.git")
	t.Logf("endpoint: %+v", e)
	assert.Equal(t, "git@git-host.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "repo.git", e.SSHMetadata.Path)

	e = lfshttp.EndpointFromBareSshUrl("git@git-host.com/should-be-a-colon.git")
	t.Logf("endpoint: %+v", e)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)

	e = lfshttp.EndpointFromBareSshUrl("-oProxyCommand=gnome-calculator")
	t.Logf("endpoint: %+v", e)
	assert.Equal(t, "", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "", e.SSHMetadata.Path)

	e = lfshttp.EndpointFromBareSshUrl("git@git-host.com:-oProxyCommand=gnome-calculator")
	t.Logf("endpoint: %+v", e)
	assert.Equal(t, "git@git-host.com", e.SSHMetadata.UserAndHost)
	assert.Equal(t, "-oProxyCommand=gnome-calculator", e.SSHMetadata.Path)
}
