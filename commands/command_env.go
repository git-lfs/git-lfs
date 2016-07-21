package commands

import (
	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	envCmd = &cobra.Command{
		Use: "env",
		Run: envCommand,
	}
)

func envCommand(cmd *cobra.Command, args []string) {
	config.ShowConfigWarnings = true
	endpoint := Config.Endpoint("download")

	gitV, err := git.Config.Version()
	if err != nil {
		gitV = "Error getting git version: " + err.Error()
	}

	Print(config.VersionDesc)
	Print(gitV)
	Print("")

	if len(endpoint.Url) > 0 {
		Print("Endpoint=%s (auth=%s)", endpoint.Url, Config.EndpointAccess(endpoint))
		if len(endpoint.SshUserAndHost) > 0 {
			Print("  SSH=%s:%s", endpoint.SshUserAndHost, endpoint.SshPath)
		}
	}

	for _, remote := range Config.Remotes() {
		remoteEndpoint := Config.RemoteEndpoint(remote, "download")
		Print("Endpoint (%s)=%s (auth=%s)", remote, remoteEndpoint.Url, Config.EndpointAccess(remoteEndpoint))
		if len(remoteEndpoint.SshUserAndHost) > 0 {
			Print("  SSH=%s:%s", remoteEndpoint.SshUserAndHost, remoteEndpoint.SshPath)
		}
	}

	for _, env := range lfs.Environ() {
		Print(env)
	}

	for _, key := range []string{"filter.lfs.smudge", "filter.lfs.clean"} {
		value, _ := Config.GitConfig(key)
		Print("git config %s = %q", key, value)
	}
}

func init() {
	RootCmd.AddCommand(envCmd)
}
