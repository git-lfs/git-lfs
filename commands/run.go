package commands

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tr"
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
	tr.InitializeLocale()

	root := NewCommand("git-lfs", gitlfsCommand)
	root.PreRun = nil

	completionCmd := &cobra.Command{
		Use:                   "completion [bash|fish|zsh]",
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "fish", "zsh"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				completion := new(bytes.Buffer)
				cmd.Root().GenBashCompletionV2(completion, false)

				// this is needed for git bash completion to pick up the completion for the subcommand
				completionSource := []byte(`    local out directive
    __git-lfs_get_completion_results
`)
				completionReplace := []byte(`    if [[ ${words[0]} == "git" && ${words[1]} == "lfs" ]]; then
        words=("git-lfs" "${words[@]:2:${#words[@]}-2}")
        __git-lfs_debug "Rewritten words[*]: ${words[*]},"
    fi

    local out directive
    __git-lfs_get_completion_results
`)
				newCompletion := bytes.NewBuffer(bytes.Replace(completion.Bytes(), completionSource, completionReplace, 1))
				newCompletion.WriteString("_git_lfs() { __start_git-lfs; }\n")

				newCompletion.WriteTo(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, false)
			case "zsh":
				completion := new(bytes.Buffer)
				cmd.Root().GenZshCompletionNoDesc(completion)

				// this is needed for git zsh completion to use the right command for completion
				completionSource := []byte(`    requestComp="${words[1]} __completeNoDesc ${words[2,-1]}"`)
				completionReplace := []byte(`    requestComp="git-${words[1]#*git-} __completeNoDesc ${words[2,-1]}"`)
				newCompletion := bytes.NewBuffer(bytes.Replace(completion.Bytes(), completionSource, completionReplace, 1))

				newCompletion.WriteTo(os.Stdout)
			}
		},
	}

	root.AddCommand(completionCmd)

	// Set up help/usage funcs based on manpage text
	helpcmd := &cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Long: `Help provides help for any command in the application.
Simply type ` + root.Name() + ` help [path to command] for full details.`,

		Run: func(c *cobra.Command, args []string) {
			cmd, _, e := c.Root().Find(args)
			// In the case of "git lfs help config" or "git lfs help
			// faq", pretend the last arg was "help" so our command
			// lookup succeeds, since cmd will be ignored in
			// helpCommand().
			if e != nil && (args[0] == "config" || args[0] == "faq") {
				cmd, _, e = c.Root().Find([]string{"help"})
			}
			if cmd == nil || e != nil {
				c.Println(tr.Tr.Get("Unknown help topic %#q", args))
				c.Root().Usage()
			} else {
				c.HelpFunc()(cmd, args)
			}
		},
	}

	root.SetHelpCommand(helpcmd)

	root.SetHelpTemplate("{{.UsageString}}")
	root.SetHelpFunc(helpCommand)
	root.SetUsageFunc(usageCommand)

	root.Flags().BoolVarP(&rootVersion, "version", "v", false, "")

	canonicalizeEnvironment()

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
	if commandName == "--help" {
		commandName = "git-lfs"
	}
	if txt, ok := ManPages[commandName]; ok {
		fmt.Println(strings.TrimSpace(txt))
	} else {
		fmt.Println(tr.Tr.Get("Sorry, no usage text found for %q", commandName))
	}
}

func setupHTTPLogger(cmd *cobra.Command, args []string) {
	if len(os.Getenv("GIT_LOG_STATS")) < 1 {
		return
	}

	logBase := filepath.Join(cfg.LocalLogDir(), "http")
	if err := tools.MkdirAll(logBase, cfg); err != nil {
		fmt.Fprintln(os.Stderr, tr.Tr.Get("Error logging HTTP stats: %s", err))
		return
	}

	logFile := fmt.Sprintf("http-%d.log", time.Now().Unix())
	file, err := os.Create(filepath.Join(logBase, logFile))
	if err != nil {
		fmt.Fprintln(os.Stderr, tr.Tr.Get("Error logging HTTP stats: %s", err))
	} else {
		getAPIClient().LogHTTPStats(file)
	}
}
