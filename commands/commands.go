package gitmedia

import (
	"fmt"
	"flag"
	"os"
)

var commands = make(map[string]func(*Command) RunnableCommand)

func Run() {
	runcmd := true
	subcommand := SubCommand(1)

	if subcommand == "help" {
		runcmd = false
		subcommand = SubCommand(2)
	}

	cmdcb, ok := commands[subcommand]
	if ok {
		cmd := cmdcb(NewCommand(subcommand))
		cmd.Setup()

		if runcmd {
			cmd.Parse()
			cmd.Run()
		} else {
			cmd.Usage()
		}
	} else {
		missingCommand(subcommand)
	}
}

func SubCommand(pos int) string {
	if len(os.Args) < (pos + 1) {
		return "version"
	} else {
		return os.Args[pos]
	}
}

func NewCommand(name string) *Command {
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[2:]
	}

	return &Command{name, flag.NewFlagSet(os.Args[0], flag.ExitOnError), args}
}

type RunnableCommand interface {
	Setup()
	Parse()
	Run()
	Usage()
}

type Command struct {
	Name string
	FlagSet *flag.FlagSet
	Args    []string
}

func (c *Command) Usage() {
	fmt.Printf("git-media %s\n", c.Name)
	c.FlagSet.PrintDefaults()
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
