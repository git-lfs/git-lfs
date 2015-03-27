package commands

import (
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	envCmd = &cobra.Command{
		Use:   "env",
		Short: "Show the current environment",
		Run:   envCommand,
	}
)

func envCommand(cmd *cobra.Command, args []string) {
	config := lfs.Config

	endpoint := config.Endpoint()

	if len(endpoint.Url) > 0 {
		Print("Endpoint=%s", endpoint.Url)
		if len(endpoint.SshUserAndHost) > 0 {
			Print("  SSH=%s:%s", endpoint.SshUserAndHost, endpoint.SshPath)
		}
	}

	for _, remote := range config.Remotes() {
		remoteEndpoint := config.RemoteEndpoint(remote)
		Print("Endpoint (%s)=%s", remote, remoteEndpoint.Url)
		if len(endpoint.SshUserAndHost) > 0 {
			Print("  SSH=%s:%s", endpoint.SshUserAndHost, endpoint.SshPath)
		}
	}

	for _, env := range lfs.Environ() {
		Print(env)
	}
}

func init() {
	RootCmd.AddCommand(envCmd)
}
