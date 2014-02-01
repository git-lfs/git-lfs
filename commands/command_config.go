package gitmedia

import (
	core ".."
)

type ConfigCommand struct {
	*Command
}

func (c *ConfigCommand) Run() {
	config := core.Config

	if endpoint := config.Endpoint(); len(endpoint) > 0 {
		core.Print("Endpoint=%s", endpoint)
	}

	for _, remote := range config.Remotes() {
		core.Print("Endpoint (%s)=%s", remote, config.RemoteEndpoint(remote))
	}

	for _, env := range core.Environ() {
		core.Print(env)
	}
}

func init() {
	registerCommand("config", func(c *Command) RunnableCommand {
		return &ConfigCommand{Command: c}
	})
}
