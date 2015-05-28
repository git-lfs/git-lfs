// Copyright © 2013 Steve Francia <spf@spf13.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Commands similar to git, go tools and other modern CLI tools
// inspired by go, go-Commander, gh and subcommand

package cobra

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	flag "github.com/github/git-lfs/vendor/_nuts/github.com/spf13/pflag"
)

// Command is just that, a command for your application.
// eg.  'go run' ... 'run' is the command. Cobra requires
// you to define the usage and description as part of your command
// definition to ensure usability.
type Command struct {
	// Name is the command name, usually the executable's name.
	name string
	// The one-line usage message.
	Use string
	// The short description shown in the 'help' output.
	Short string
	// The long message shown in the 'help <this-command>' output.
	Long string
	// Set of flags specific to this command.
	flags *flag.FlagSet
	// Set of flags children commands will inherit
	pflags *flag.FlagSet
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(cmd *Command, args []string)
	// Commands is the list of commands supported by this program.
	commands []*Command
	// Parent Command for this command
	parent *Command
	// max lengths of commands' string lengths for use in padding
	commandsMaxUseLen         int
	commandsMaxCommandPathLen int

	flagErrorBuf *bytes.Buffer

	args          []string
	output        *io.Writer               // nil means stderr; use Out() method instead
	usageFunc     func(*Command) error     // Usage can be defined by application
	usageTemplate string                   // Can be defined by Application
	helpTemplate  string                   // Can be defined by Application
	helpFunc      func(*Command, []string) // Help can be defined by application
	helpCommand   *Command                 // The help command
	helpFlagVal   bool
}

// os.Args[1:] by default, if desired, can be overridden
// particularly useful when testing.
func (c *Command) SetArgs(a []string) {
	c.args = a
}

func (c *Command) Out() io.Writer {
	if c.output != nil {
		return *c.output
	}

	if c.HasParent() {
		return c.parent.Out()
	} else {
		return os.Stderr
	}
}

// SetOutput sets the destination for usage and error messages.
// If output is nil, os.Stderr is used.
func (c *Command) SetOutput(output io.Writer) {
	c.output = &output
}

// Usage can be defined by application
func (c *Command) SetUsageFunc(f func(*Command) error) {
	c.usageFunc = f
}

// Can be defined by Application
func (c *Command) SetUsageTemplate(s string) {
	c.usageTemplate = s
}

// Can be defined by Application
func (c *Command) SetHelpFunc(f func(*Command, []string)) {
	c.helpFunc = f
}

func (c *Command) SetHelpCommand(cmd *Command) {
	c.helpCommand = cmd
}

// Can be defined by Application
func (c *Command) SetHelpTemplate(s string) {
	c.helpTemplate = s
}

func (c *Command) UsageFunc() (f func(*Command) error) {
	if c.usageFunc != nil {
		return c.usageFunc
	}

	if c.HasParent() {
		return c.parent.UsageFunc()
	} else {
		return func(c *Command) error {
			err := tmpl(c.Out(), c.UsageTemplate(), c)
			return err
		}
	}
}
func (c *Command) HelpFunc() func(*Command, []string) {
	if c.helpFunc != nil {
		return c.helpFunc
	}

	if c.HasParent() {
		return c.parent.HelpFunc()
	} else {
		return func(c *Command, args []string) {
			if len(args) == 0 {
				// Help called without any topic, calling on root
				c.Root().Usage()
				return
			}

			cmd, _, e := c.Root().Find(args)
			if cmd == nil || e != nil {
				c.Printf("Unknown help topic %#q.", args)

				c.Root().Usage()
			} else {
				err := cmd.Help()
				if err != nil {
					c.Println(err)
				}
			}
		}
	}
}

var minUsagePadding int = 25

func (c *Command) UsagePadding() int {
	if c.parent == nil || minUsagePadding > c.parent.commandsMaxUseLen {
		return minUsagePadding
	} else {
		return c.parent.commandsMaxUseLen
	}
}

var minCommandPathPadding int = 11

func (c *Command) CommandPathPadding() int {
	if c.parent == nil || minCommandPathPadding > c.parent.commandsMaxCommandPathLen {
		return minCommandPathPadding
	} else {
		return c.parent.commandsMaxCommandPathLen
	}
}

func (c *Command) UsageTemplate() string {
	if c.usageTemplate != "" {
		return c.usageTemplate
	}

	if c.HasParent() {
		return c.parent.UsageTemplate()
	} else {
		return `{{ $cmd := . }}
Usage: {{if .Runnable}}
  {{.UseLine}}{{if .HasFlags}} [flags]{{end}}{{end}}{{if .HasSubCommands}}
  {{ .CommandPath}} [command]{{end}}
{{ if .HasSubCommands}}
Available Commands: {{range .Commands}}{{if .Runnable}}
  {{rpad .Use .UsagePadding }} {{.Short}}{{end}}{{end}}
{{end}}
{{ if .HasFlags}} Available Flags:
{{.Flags.FlagUsages}}{{end}}{{if .HasParent}}{{if and (gt .Commands 0) (gt .Parent.Commands 1) }}
Additional help topics: {{if gt .Commands 0 }}{{range .Commands}}{{if not .Runnable}} {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if gt .Parent.Commands 1 }}{{range .Parent.Commands}}{{if .Runnable}}{{if not (eq .Name $cmd.Name) }}{{end}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{end}}
{{end}}
Use "{{.Root.Name}} help [command]" for more information about that command.
`
	}
}

func (c *Command) HelpTemplate() string {
	if c.helpTemplate != "" {
		return c.helpTemplate
	}

	if c.HasParent() {
		return c.parent.HelpTemplate()
	} else {
		return `{{.Long | trim}}
{{if .Runnable}}{{.UsageString}}{{end}}
`
	}
}

// Really only used when casting a command to a commander
func (c *Command) resetChildrensParents() {
	for _, x := range c.commands {
		x.parent = c
	}
}

func stripFlags(args []string) []string {
	if len(args) < 1 {
		return args
	}

	commands := []string{}

	inQuote := false
	for _, y := range args {
		if !inQuote {
			switch {
			case strings.HasPrefix(y, "\""):
				inQuote = true
			case strings.Contains(y, "=\""):
				inQuote = true
			case !strings.HasPrefix(y, "-"):
				commands = append(commands, y)
			}
		}

		if strings.HasSuffix(y, "\"") && !strings.HasSuffix(y, "\\\"") {
			inQuote = false
		}
	}

	return commands
}

func argsMinusX(args []string, x string) []string {
	newargs := []string{}

	for _, y := range args {
		if x != y {
			newargs = append(newargs, y)
		}
	}
	return newargs
}

// find the target command given the args and command tree
// Meant to be run on the highest node. Only searches down.
func (c *Command) Find(arrs []string) (*Command, []string, error) {
	if c == nil {
		return nil, nil, fmt.Errorf("Called find() on a nil Command")
	}

	if len(arrs) == 0 {
		return c.Root(), arrs, nil
	}

	var innerfind func(*Command, []string) (*Command, []string)

	innerfind = func(c *Command, args []string) (*Command, []string) {
		if len(args) > 0 && c.HasSubCommands() {
			argsWOflags := stripFlags(args)
			if len(argsWOflags) > 0 {
				matches := make([]*Command, 0)
				for _, cmd := range c.commands {
					if cmd.Name() == argsWOflags[0] { // exact name match
						return innerfind(cmd, argsMinusX(args, cmd.Name()))
					} else if strings.HasPrefix(cmd.Name(), argsWOflags[0]) { // prefix match
						matches = append(matches, cmd)
					}
				}

				// only accept a single prefix match - multiple matches would be ambiguous
				if len(matches) == 1 {
					return innerfind(matches[0], argsMinusX(args, argsWOflags[0]))
				}
			}
		}

		return c, args
	}

	commandFound, a := innerfind(c, arrs)

	// if commander returned and not appropriately matched return nil & error
	if commandFound.Name() == c.Name() && !strings.HasPrefix(commandFound.Name(), arrs[0]) {
		return nil, a, fmt.Errorf("unknown command %q\nRun 'help' for usage.\n", a[0])
	}

	return commandFound, a, nil
}

func (c *Command) Root() *Command {
	var findRoot func(*Command) *Command

	findRoot = func(x *Command) *Command {
		if x.HasParent() {
			return findRoot(x.parent)
		} else {
			return x
		}
	}

	return findRoot(c)
}

// execute the command determined by args and the command tree
func (c *Command) findAndExecute(args []string) (err error) {

	cmd, a, e := c.Find(args)
	if e != nil {
		return e
	}
	return cmd.execute(a)
}

func (c *Command) execute(a []string) (err error) {
	if c == nil {
		return fmt.Errorf("Called Execute() on a nil Command")
	}

	err = c.ParseFlags(a)

	if err != nil {
		return err
	} else {
		// If help is called, regardless of other flags, we print that
		if c.helpFlagVal {
			c.Help()
			return nil
		}

		c.preRun()
		argWoFlags := c.Flags().Args()
		c.Run(c, argWoFlags)
		return nil
	}
}

func (c *Command) preRun() {
	for _, x := range initializers {
		x()
	}
}

func (c *Command) errorMsgFromParse() string {
	s := c.flagErrorBuf.String()

	x := strings.Split(s, "\n")

	if len(x) > 0 {
		return x[0]
	} else {
		return ""
	}
}

// Call execute to use the args (os.Args[1:] by default)
// and run through the command tree finding appropriate matches
// for commands and then corresponding flags.
func (c *Command) Execute() (err error) {

	// Regardless of what command execute is called on, run on Root only
	if c.HasParent() {
		return c.Root().Execute()
	}

	// initialize help as the last point possible to allow for user
	// overriding
	c.initHelp()

	var args []string

	if len(c.args) == 0 {
		args = os.Args[1:]
	} else {
		args = c.args
	}

	if len(args) == 0 {
		// Only the executable is called and the root is runnable, run it
		if c.Runnable() {
			err = c.execute([]string(nil))
		} else {
			c.Usage()
		}
	} else {
		err = c.findAndExecute(args)
	}

	// Now handle the case where the root is runnable and only flags are provided
	if err != nil && c.Runnable() {
		// This is pretty much a custom version of the *Command.execute method
		// with a few differences because it's the final command (no fall back)
		e := c.ParseFlags(args)
		if e != nil {
			// Flags parsing had an error.
			// If an error happens here, we have to report it to the user
			c.Println(c.errorMsgFromParse())
			c.Usage()
			return e
		} else {
			// If help is called, regardless of other flags, we print that
			if c.helpFlagVal {
				c.Help()
				return nil
			}

			argWoFlags := c.Flags().Args()
			if len(argWoFlags) > 0 {
				// If there are arguments (not flags) one of the earlier
				// cases should have caught it.. It means invalid usage
				// print the usage
				c.Usage()
			} else {
				// Only flags left... Call root.Run
				c.preRun()
				c.Run(c, argWoFlags)
				err = nil
			}
		}
	}

	if err != nil {
		c.Println("Error:", err.Error())
		c.Printf("%v: invalid command %#q\n", c.Root().Name(), os.Args[1:])
		c.Printf("Run '%v help' for usage\n", c.Root().Name())
	}

	return
}

func (c *Command) initHelp() {
	if c.helpCommand == nil {
		c.helpCommand = &Command{
			Use:   "help [command]",
			Short: "Help about any command",
			Long: `Help provides help for any command in the application.
    Simply type ` + c.Name() + ` help [path to command] for full details.`,
			Run: c.HelpFunc(),
		}
	}
	c.AddCommand(c.helpCommand)
}

// Used for testing
func (c *Command) ResetCommands() {
	c.commands = nil
}

func (c *Command) Commands() []*Command {
	return c.commands
}

// Add one or many commands as children of this
func (c *Command) AddCommand(cmds ...*Command) {
	for i, x := range cmds {
		if cmds[i] == c {
			panic("Command can't be a child of itself")
		}
		cmds[i].parent = c
		// update max lengths
		usageLen := len(x.Use)
		if usageLen > c.commandsMaxUseLen {
			c.commandsMaxUseLen = usageLen
		}
		commandPathLen := len(x.CommandPath())
		if commandPathLen > c.commandsMaxCommandPathLen {
			c.commandsMaxCommandPathLen = commandPathLen
		}
		c.commands = append(c.commands, x)
	}
}

// Convenience method to Print to the defined output
func (c *Command) Print(i ...interface{}) {
	fmt.Fprint(c.Out(), i...)
}

// Convenience method to Println to the defined output
func (c *Command) Println(i ...interface{}) {
	str := fmt.Sprintln(i...)
	c.Print(str)
}

// Convenience method to Printf to the defined output
func (c *Command) Printf(format string, i ...interface{}) {
	str := fmt.Sprintf(format, i...)
	c.Print(str)
}

// Output the usage for the command
// Used when a user provides invalid input
// Can be defined by user by overriding UsageFunc
func (c *Command) Usage() error {
	c.mergePersistentFlags()
	err := c.UsageFunc()(c)
	return err
}

// Output the help for the command
// Used when a user calls help [command]
// by the default HelpFunc in the commander
func (c *Command) Help() error {
	c.mergePersistentFlags()
	err := tmpl(c.Out(), c.HelpTemplate(), c)
	return err
}

func (c *Command) UsageString() string {
	tmpOutput := c.output
	bb := new(bytes.Buffer)
	c.SetOutput(bb)
	c.Usage()
	c.output = tmpOutput
	return bb.String()
}

// The full path to this command
func (c *Command) CommandPath() string {
	str := c.Name()
	x := c
	for x.HasParent() {
		str = x.parent.Name() + " " + str
		x = x.parent
	}
	return str
}

//The full usage for a given command (including parents)
func (c *Command) UseLine() string {
	str := ""
	if c.HasParent() {
		str = c.parent.CommandPath() + " "
	}
	return str + c.Use
}

// For use in determining which flags have been assigned to which commands
// and which persist
func (c *Command) DebugFlags() {
	c.Println("DebugFlags called on", c.Name())
	var debugflags func(*Command)

	debugflags = func(x *Command) {
		if x.HasFlags() || x.HasPersistentFlags() {
			c.Println(x.Name())
		}
		if x.HasFlags() {
			x.flags.VisitAll(func(f *flag.Flag) {
				if x.HasPersistentFlags() {
					if x.persistentFlag(f.Name) == nil {
						c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [L]")
					} else {
						c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [LP]")
					}
				} else {
					c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [L]")
				}
			})
		}
		if x.HasPersistentFlags() {
			x.pflags.VisitAll(func(f *flag.Flag) {
				if x.HasFlags() {
					if x.flags.Lookup(f.Name) == nil {
						c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [P]")
					}
				} else {
					c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [P]")
				}
			})
		}
		c.Println(x.flagErrorBuf)
		if x.HasSubCommands() {
			for _, y := range x.commands {
				debugflags(y)
			}
		}
	}

	debugflags(c)
}

// Name returns the command's name: the first word in the use line.
func (c *Command) Name() string {
	if c.name != "" {
		return c.name
	}
	name := c.Use
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

// Determine if the command is itself runnable
func (c *Command) Runnable() bool {
	return c.Run != nil
}

// Determine if the command has children commands
func (c *Command) HasSubCommands() bool {
	return len(c.commands) > 0
}

// Determine if the command is a child command
func (c *Command) HasParent() bool {
	return c.parent != nil
}

// Get the Commands FlagSet
func (c *Command) Flags() *flag.FlagSet {
	if c.flags == nil {
		c.flags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.flags.SetOutput(c.flagErrorBuf)
		c.flags.BoolVar(&c.helpFlagVal, "help", false, "help for "+c.Name())
	}
	return c.flags
}

// Get the Commands Persistent FlagSet
func (c *Command) PersistentFlags() *flag.FlagSet {
	if c.pflags == nil {
		c.pflags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.pflags.SetOutput(c.flagErrorBuf)
	}
	return c.pflags
}

// For use in testing
func (c *Command) ResetFlags() {
	c.flagErrorBuf = new(bytes.Buffer)
	c.flagErrorBuf.Reset()
	c.flags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	c.flags.SetOutput(c.flagErrorBuf)
	c.pflags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	c.pflags.SetOutput(c.flagErrorBuf)
}

// Does the command contain flags (local not persistent)
func (c *Command) HasFlags() bool {
	return c.Flags().HasFlags()
}

// Does the command contain persistent flags
func (c *Command) HasPersistentFlags() bool {
	return c.PersistentFlags().HasFlags()
}

// Climbs up the command tree looking for matching flag
func (c *Command) Flag(name string) (flag *flag.Flag) {
	flag = c.Flags().Lookup(name)

	if flag == nil {
		flag = c.persistentFlag(name)
	}

	return
}

// recursively find matching persistent flag
func (c *Command) persistentFlag(name string) (flag *flag.Flag) {
	if c.HasPersistentFlags() {
		flag = c.PersistentFlags().Lookup(name)
	}

	if flag == nil && c.HasParent() {
		flag = c.parent.persistentFlag(name)
	}
	return
}

// Parses persistent flag tree & local flags
func (c *Command) ParseFlags(args []string) (err error) {
	c.mergePersistentFlags()
	err = c.Flags().Parse(args)

	// The upstream library adds spaces to the error
	// response regardless of success.
	// Handling it here until fixing upstream
	if len(strings.TrimSpace(c.flagErrorBuf.String())) > 1 {
		return fmt.Errorf("%s", c.flagErrorBuf.String())
	}

	//always return nil because upstream library is inconsistent & we always check the error buffer anyway
	return nil
}

func (c *Command) mergePersistentFlags() {
	var rmerge func(x *Command)

	rmerge = func(x *Command) {
		if x.HasPersistentFlags() {
			x.PersistentFlags().VisitAll(func(f *flag.Flag) {
				if c.Flags().Lookup(f.Name) == nil {
					c.Flags().AddFlag(f)
				}
			})
		}
		if x.HasParent() {
			rmerge(x.parent)
		}
	}

	rmerge(c)
}

func (c *Command) Parent() *Command {
	return c.parent
}
