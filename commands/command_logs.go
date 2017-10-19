package commands

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/spf13/cobra"
)

func logsCommand(cmd *cobra.Command, args []string) {
	for _, path := range sortedLogs() {
		Print(path)
	}
}

func logsLastCommand(cmd *cobra.Command, args []string) {
	logs := sortedLogs()
	if len(logs) < 1 {
		Print("No logs to show")
		return
	}

	logsShowCommand(cmd, logs[len(logs)-1:])
}

func logsShowCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		Print("Supply a log name.")
		return
	}

	name := args[0]
	by, err := ioutil.ReadFile(filepath.Join(cfg.LocalLogDir(), name))
	if err != nil {
		Exit("Error reading log: %s", name)
	}

	Debug("Reading log: %s", name)
	os.Stdout.Write(by)
}

func logsClearCommand(cmd *cobra.Command, args []string) {
	err := os.RemoveAll(cfg.LocalLogDir())
	if err != nil {
		Panic(err, "Error clearing %s", cfg.LocalLogDir())
	}

	Print("Cleared %s", cfg.LocalLogDir())
}

func logsBoomtownCommand(cmd *cobra.Command, args []string) {
	Debug("Debug message")
	err := errors.Wrapf(errors.New("Inner error message!"), "Error")
	Panic(err, "Welcome to Boomtown")
	Debug("Never seen")
}

func sortedLogs() []string {
	fileinfos, err := ioutil.ReadDir(cfg.LocalLogDir())
	if err != nil {
		return []string{}
	}

	names := make([]string, 0, len(fileinfos))
	for _, info := range fileinfos {
		if info.IsDir() {
			continue
		}
		names = append(names, info.Name())
	}

	return names
}

func init() {
	RegisterCommand("logs", logsCommand, func(cmd *cobra.Command) {
		cmd.AddCommand(
			NewCommand("last", logsLastCommand),
			NewCommand("show", logsShowCommand),
			NewCommand("clear", logsClearCommand),
			NewCommand("boomtown", logsBoomtownCommand),
		)
	})
}
