package gitmedia

import (
	core ".."
	"fmt"
)

type ConfigCommand struct {
	*Command
}

func (c *ConfigCommand) Run() {
	config := core.Config()
	fmt.Printf("Endpoint: %s\n", config.Endpoint)
}

func init() {
	registerCommand("config", func(c *Command) RunnableCommand {
		return &ConfigCommand{Command: c}
	})
}
