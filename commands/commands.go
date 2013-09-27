package gitmedia

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var commands = make(map[string]func(*Command) RunnableCommand)

func Run() {
	runcmd := true
	subname := SubCommand(1)

	if subname == "help" {
		runcmd = false
		subname = SubCommand(2)
	}

	cmd := NewCommand(filepath.Base(os.Args[0]), subname)
	cmdcb, ok := commands[subname]
	if ok {
		subcmd := cmdcb(cmd)
		subcmd.Setup()

		if runcmd {
			subcmd.Parse()
			subcmd.Run()
		} else {
			subcmd.Usage()
		}
	} else {
		missingCommand(cmd, subname)
	}
}

func SubCommand(pos int) string {
	if len(os.Args) < (pos + 1) {
		return "version"
	} else {
		return os.Args[pos]
	}
}

func NewCommand(name, subname string) *Command {
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[2:]
	}

	return &Command{name, subname, flag.NewFlagSet(os.Args[0], flag.ExitOnError), args}
}

type RunnableCommand interface {
	Setup()
	Parse()
	Run()
	Usage()
}

type Command struct {
	Name       string
	SubCommand string
	FlagSet    *flag.FlagSet
	Args       []string
}

func (c *Command) Usage() {
	fmt.Printf("usage: %s %s\n", c.Name, c.SubCommand)
	c.FlagSet.PrintDefaults()
}

func (c *Command) Parse() {
	c.FlagSet.Parse(c.Args)
}

func (c *Command) Setup() {}
func (c *Command) Run()   {}

func registerCommand(name string, cmdcb func(*Command) RunnableCommand) {
	commands[name] = cmdcb
}

func missingCommand(cmd *Command, subname string) {
	fmt.Printf("%s: '%s' is not a %s command.  See %s help.\n",
		cmd.Name, subname, cmd.Name, cmd.Name)
}
