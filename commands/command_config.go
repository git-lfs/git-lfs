package commands

import (
	"github.com/github/git-media/gitmedia"
)

type ConfigCommand struct {
	*Command
}

func (c *ConfigCommand) Run() {
	config := gitmedia.Config

	if endpoint := config.Endpoint(); len(endpoint) > 0 {
		gitmedia.Print("Endpoint=%s", endpoint)
	}

	for _, remote := range config.Remotes() {
		gitmedia.Print("Endpoint (%s)=%s", remote, config.RemoteEndpoint(remote))
	}

	for _, env := range gitmedia.Environ() {
		gitmedia.Print(env)
	}
}

func init() {
	registerCommand("config", func(c *Command) RunnableCommand {
		return &ConfigCommand{Command: c}
	})
}
