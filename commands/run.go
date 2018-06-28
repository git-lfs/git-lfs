package commands

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/git-lfs/git-lfs/config"
	"github.com/spf13/cobra"
)

var (
	commandFuncs []func() *cobra.Command
	commandMu    sync.Mutex

	rootVersion bool
)

// NewCommand creates a new 'git-lfs' sub command, given a command name and
// command run function.
//
// Each command will initialize the local storage ('.git/lfs') directory when
// run, unless the PreRun hook is set to nil.
func NewCommand(name string, runFn func(*cobra.Command, []string)) *cobra.Command {
	return &cobra.Command{Use: name, Run: runFn, PreRun: setupHTTPLogger}
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
//
// It returns an exit code.
func Run() int {
	log.SetOutput(ErrorWriter)

	root := NewCommand("git-lfs", gitlfsCommand)
	root.PreRun = nil

	// Set up help/usage funcs based on manpage text
	root.SetHelpTemplate("{{.UsageString}}")
	root.SetHelpFunc(helpCommand)
	root.SetUsageFunc(usageCommand)

	root.Flags().BoolVarP(&rootVersion, "version", "v", false, "")

	cfg = config.New()

	for _, f := range commandFuncs {
		if cmd := f(); cmd != nil {
			root.AddCommand(cmd)
		}
	}

	err := root.Execute()
	closeAPIClient()

	if err != nil {
		return 127
	}
	return 0
}

func gitlfsCommand(cmd *cobra.Command, args []string) {
	versionCommand(cmd, args)
	if !rootVersion {
		cmd.Usage()
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
		fmt.Fprintf(os.Stdout, "%s\n", strings.TrimSpace(txt))
	} else {
		fmt.Fprintf(os.Stdout, "Sorry, no usage text found for %q\n", commandName)
	}
}

func setupHTTPLogger(cmd *cobra.Command, args []string) {
	if len(os.Getenv("GIT_LOG_STATS")) < 1 {
		return
	}

	logBase := filepath.Join(cfg.LocalLogDir(), "http")
	if err := os.MkdirAll(logBase, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error logging http stats: %s\n", err)
		return
	}

	logFile := fmt.Sprintf("http-%d.log", time.Now().Unix())
	file, err := os.Create(filepath.Join(logBase, logFile))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error logging http stats: %s\n", err)
	} else {
		getAPIClient().LogHTTPStats(file)
	}
}
