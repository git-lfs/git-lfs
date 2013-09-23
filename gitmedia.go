package gitmedia

import (
	"fmt"
)

const Version = "0.0.1"

var commands = make(map[string]*Command)

func Run() {
	subcommand := SubCommand()
	cmd, ok := commands[subcommand]
	if ok {
		cmd.Run()
	} else {
		missingCommand(subcommand)
	}
}

func registerCommand(name string, cmd *Command) {
	commands[name] = cmd
}

func missingCommand(cmd string) {
	fmt.Printf("git-media: '%s' is not a git-media command.  See git-media help.\n", cmd)
}
