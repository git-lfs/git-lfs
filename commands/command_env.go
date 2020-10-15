package commands

import (
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/spf13/cobra"
)

func envCommand(cmd *cobra.Command, args []string) {
	config.ShowConfigWarnings = true

	gitV, err := git.Version()
	if err != nil {
		gitV = "Error getting git version: " + err.Error()
	}

	Print(config.VersionDesc)
	Print(gitV)
	Print("")

	defaultRemote := ""
	if cfg.IsDefaultRemote() {
		defaultRemote = cfg.Remote()
		endpoint := getAPIClient().Endpoints.Endpoint("download", defaultRemote)
		if len(endpoint.Url) > 0 {
			access := getAPIClient().Endpoints.AccessFor(endpoint.Url)
			Print("Endpoint=%s (auth=%s)", endpoint.Url, access.Mode())
			if len(endpoint.SshUserAndHost) > 0 {
				Print("  SSH=%s:%s", endpoint.SshUserAndHost, endpoint.SshPath)
			}
		}
	}

	for _, remote := range cfg.Remotes() {
		if remote == defaultRemote {
			continue
		}
		remoteEndpoint := getAPIClient().Endpoints.Endpoint("download", remote)
		remoteAccess := getAPIClient().Endpoints.AccessFor(remoteEndpoint.Url)
		Print("Endpoint (%s)=%s (auth=%s)", remote, remoteEndpoint.Url, remoteAccess.Mode())
		if len(remoteEndpoint.SshUserAndHost) > 0 {
			Print("  SSH=%s:%s", remoteEndpoint.SshUserAndHost, remoteEndpoint.SshPath)
		}
	}

	for _, env := range lfs.Environ(cfg, getTransferManifest(), oldEnv) {
		Print(env)
	}

	for _, key := range []string{"filter.lfs.process", "filter.lfs.smudge", "filter.lfs.clean"} {
		value, _ := cfg.Git.Get(key)
		Print("git config %s = %q", key, value)
	}
}

func init() {
	RegisterCommand("env", envCommand, nil)
}
