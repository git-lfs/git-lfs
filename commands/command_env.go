package commands

import (
	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/spf13/cobra"
)

func envCommand(cmd *cobra.Command, args []string) {
	config.ShowConfigWarnings = true

	gitV, err := git.Version()
	if err != nil {
		gitV = tr.Tr.Get("Error getting Git version: %s", err.Error())
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
			if len(endpoint.SSHMetadata.UserAndHost) > 0 {
				Print("  SSH=%s:%s", endpoint.SSHMetadata.UserAndHost, endpoint.SSHMetadata.Path)
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
		if len(remoteEndpoint.SSHMetadata.UserAndHost) > 0 {
			Print("  SSH=%s:%s", remoteEndpoint.SSHMetadata.UserAndHost, remoteEndpoint.SSHMetadata.Path)
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
