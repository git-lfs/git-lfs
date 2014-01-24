package gitmedia

import (
	core ".."
)

type ConfigCommand struct {
	*Command
}

func (c *ConfigCommand) Run() {
	config := core.Config()
	core.Print("Endpoint=%s", config.Endpoint)
	for _, env := range core.Environ() {
		core.Print(env)
	}
}

func init() {
	registerCommand("config", func(c *Command) RunnableCommand {
		return &ConfigCommand{Command: c}
	})
}
