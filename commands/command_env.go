package commands

import (
	"github.com/github/git-media/gitmedia"
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
	config := gitmedia.Config

	if endpoint := config.Endpoint(); len(endpoint) > 0 {
		Print("Endpoint=%s", endpoint)
	}

	for _, remote := range config.Remotes() {
		Print("Endpoint (%s)=%s", remote, config.RemoteEndpoint(remote))
	}

	for _, env := range gitmedia.Environ() {
		Print(env)
	}
}

func init() {
	RootCmd.AddCommand(envCmd)
}
