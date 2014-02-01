package gitmedia

import (
	core ".."
)

type ConfigCommand struct {
	*Command
}

func (c *ConfigCommand) Run() {
	config := core.Config()

	if len(config.Endpoint) > 0 {
		core.Print("Endpoint=%s", config.Endpoint)
	} else {
		for _, remote := range config.Remotes() {
			core.Print("Endpoint (%s)=%s", remote, config.RemoteEndpoint(remote))
		}
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
