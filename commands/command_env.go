package commands

import (
	"github.com/github/git-media/gitmedia"
)

type EnvCommand struct {
	*Command
}

func (c *EnvCommand) Run() {
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
	registerCommand("env", func(c *Command) RunnableCommand {
		return &EnvCommand{Command: c}
	})
}
