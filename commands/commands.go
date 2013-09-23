package gitmedia

import (
	"fmt"
	"flag"
	"os"
)

var commands = make(map[string]func(*Command) RunnableCommand)

func Run() {
	subcommand := SubCommand()
	basecmd := NewCommand()

	cmdcb, ok := commands[subcommand]
	if ok {
		cmd := cmdcb(basecmd)
		cmd.Setup()
		cmd.Parse()
		cmd.Run()
	} else {
		missingCommand(subcommand)
	}
}

func SubCommand() string {
	if len(os.Args) < 2 {
		return "version"
	} else {
		return os.Args[1]
	}
}

func NewCommand() *Command {
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[2:]
	}

	return &Command{flag.NewFlagSet(os.Args[0], flag.ExitOnError), args}
}

type RunnableCommand interface {
	Setup()
	Parse()
	Run()
}

type Command struct {
	FlagSet *flag.FlagSet
	Args    []string
}

func (c *Command) Parse() {
	c.FlagSet.Parse(c.Args)
}

func (c *Command) Setup() {}
func (c *Command) Run() {}

func registerCommand(name string, cmdcb func(*Command) RunnableCommand) {
	commands[name] = cmdcb
}

func missingCommand(cmd string) {
	fmt.Printf("git-media: '%s' is not a git-media command.  See git-media help.\n", cmd)
}
