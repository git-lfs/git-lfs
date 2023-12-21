package commands

import (
	"os"
	"path/filepath"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/tr"
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
		Print(tr.Tr.Get("No logs to show"))
		return
	}

	logsShowCommand(cmd, logs[len(logs)-1:])
}

func logsShowCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		Print(tr.Tr.Get("Supply a log name."))
		return
	}

	name := args[0]
	by, err := os.ReadFile(filepath.Join(cfg.LocalLogDir(), name))
	if err != nil {
		Exit(tr.Tr.Get("Error reading log: %s", name))
	}

	Debug(tr.Tr.Get("Reading log: %s", name))
	os.Stdout.Write(by)
}

func logsClearCommand(cmd *cobra.Command, args []string) {
	err := os.RemoveAll(cfg.LocalLogDir())
	if err != nil {
		Panic(err, tr.Tr.Get("Error clearing %s", cfg.LocalLogDir()))
	}

	Print(tr.Tr.Get("Cleared %s", cfg.LocalLogDir()))
}

func logsBoomtownCommand(cmd *cobra.Command, args []string) {
	Debug(tr.Tr.Get("Sample debug message"))
	err := errors.Wrapf(errors.New(tr.Tr.Get("Sample wrapped error message")), tr.Tr.Get("Sample error message"))
	Panic(err, tr.Tr.Get("Sample panic message"))
}

func sortedLogs() []string {
	fileinfos, err := os.ReadDir(cfg.LocalLogDir())
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
