package commands

import (
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	envCmd = &cobra.Command{
		Use: "env",
		Run: envCommand,
	}
)

func envCommand(cmd *cobra.Command, args []string) {
	config := lfs.Config
	endpoint := config.Endpoint()

	gitV, err := git.Config.Version()
	if err != nil {
		gitV = "Error getting git version: " + err.Error()
	}

	Print(lfs.UserAgent)
	Print(gitV)
	Print("")

	if len(endpoint.Url) > 0 {
		Print("Endpoint=%s (auth=%s)", endpoint.Url, config.Access())
		if len(endpoint.SshUserAndHost) > 0 {
			Print("  SSH=%s:%s", endpoint.SshUserAndHost, endpoint.SshPath)
		}
	}

	for _, remote := range config.Remotes() {
		remoteEndpoint := config.RemoteEndpoint(remote)
		Print("Endpoint (%s)=%s (auth=%s)", remote, remoteEndpoint.Url, config.EndpointAccess(remoteEndpoint))
		if len(remoteEndpoint.SshUserAndHost) > 0 {
			Print("  SSH=%s:%s", remoteEndpoint.SshUserAndHost, remoteEndpoint.SshPath)
		}
	}

	for _, env := range lfs.Environ() {
		Print(env)
	}

	for _, key := range []string{"filter.lfs.smudge", "filter.lfs.clean"} {
		value, _ := lfs.Config.GitConfig(key)
		Print("git config %s = %q", key, value)
	}
}

func init() {
	RootCmd.AddCommand(envCmd)
}
