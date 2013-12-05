package gitmedia

import (
	core ".."
)

type ConfigCommand struct {
	*Command
}

func (c *ConfigCommand) Run() {
	config := core.Config()
	core.Print("Endpoint: %s\n", config.Endpoint)
}

func init() {
	registerCommand("config", func(c *Command) RunnableCommand {
		return &ConfigCommand{Command: c}
	})
}
