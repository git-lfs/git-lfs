package commands

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/github/git-lfs/httputil"
	"github.com/github/git-lfs/localstorage"
	"github.com/spf13/cobra"
)

var (
	commandFuncs []func() *cobra.Command
	commandMu    sync.Mutex
)

// NewCommand creates a new 'git-lfs' sub command, given a command name and
// command run function.
//
// Each command will initialize the local storage ('.git/lfs') directory when
// run, unless the PreRun hook is set to nil.
func NewCommand(name string, runFn func(*cobra.Command, []string)) *cobra.Command {
	return &cobra.Command{Use: name, Run: runFn, PreRun: commandPreRun}
}

// RegisterCommand creates a direct 'git-lfs' subcommand, given a command name,
// a command run function, and an optional callback during the command
// initialization process.
//
// The 'git-lfs' command initialization is deferred until the `commands.Run()`
// function is called. The fn callback is passed the output from NewCommand,
// and gives the caller the flexibility to customize the command by adding
// flags, tweaking command hooks, etc.
func RegisterCommand(name string, runFn func(cmd *cobra.Command, args []string), fn func(cmd *cobra.Command)) {
	commandMu.Lock()
	commandFuncs = append(commandFuncs, func() *cobra.Command {
		cmd := NewCommand(name, runFn)
		if fn != nil {
			fn(cmd)
		}
		return cmd
	})
	commandMu.Unlock()
}

// Run initializes the 'git-lfs' command and runs it with the given stdin and
// command line args.
func Run() {
	root := NewCommand("git-lfs", gitlfsCommand)
	root.PreRun = nil

	// Set up help/usage funcs based on manpage text
	root.SetHelpTemplate("{{.UsageString}}")
	root.SetHelpFunc(helpCommand)
	root.SetUsageFunc(usageCommand)

	for _, f := range commandFuncs {
		if cmd := f(); cmd != nil {
			root.AddCommand(cmd)
		}
	}

	root.Execute()
	httputil.LogHttpStats(cfg)
}

func gitlfsCommand(cmd *cobra.Command, args []string) {
	versionCommand(cmd, args)
	cmd.Usage()
}

// resolveLocalStorage implements the `func(*cobra.Command, []string)` signature
// necessary to wire it up via `cobra.Command.PreRun`. When run, this function
// will resolve the localstorage directories.
func commandPreRun(cmd *cobra.Command, args []string) {
	if err := localstorage.ResolveDirs(); err != nil {
		Exit("Init error: %v", err)
	}
}

func helpCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		printHelp("git-lfs")
	} else {
		printHelp(args[0])
	}
}

func usageCommand(cmd *cobra.Command) error {
	printHelp(cmd.Name())
	return nil
}

func printHelp(commandName string) {
	if txt, ok := ManPages[commandName]; ok {
		fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(txt))
	} else {
		fmt.Fprintf(os.Stderr, "Sorry, no usage text found for %q\n", commandName)
	}
}
