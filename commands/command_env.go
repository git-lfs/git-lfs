package commands

import (
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/spf13/cobra"
)

func envCommand(cmd *cobra.Command, args []string) {
	config.ShowConfigWarnings = true
	endpoint := cfg.Endpoint("download")

	gitV, err := git.Config.Version()
	if err != nil {
		gitV = "Error getting git version: " + err.Error()
	}

	Print(config.VersionDesc)
	Print(gitV)
	Print("")

	if len(endpoint.Url) > 0 {
		Print("Endpoint=%s (auth=%s)", endpoint.Url, cfg.EndpointAccess(endpoint))
		if len(endpoint.SshUserAndHost) > 0 {
			Print("  SSH=%s:%s", endpoint.SshUserAndHost, endpoint.SshPath)
		}
	}

	for _, remote := range cfg.Remotes() {
		remoteEndpoint := cfg.RemoteEndpoint(remote, "download")
		Print("Endpoint (%s)=%s (auth=%s)", remote, remoteEndpoint.Url, cfg.EndpointAccess(remoteEndpoint))
		if len(remoteEndpoint.SshUserAndHost) > 0 {
			Print("  SSH=%s:%s", remoteEndpoint.SshUserAndHost, remoteEndpoint.SshPath)
		}
	}

	for _, env := range lfs.Environ(cfg, getTransferManifest()) {
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
