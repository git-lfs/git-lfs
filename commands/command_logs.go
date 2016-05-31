package commands

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errutil"
	"github.com/spf13/cobra"
)

var (
	logsCmd = &cobra.Command{
		Use: "logs",
		Run: logsCommand,
	}

	logsLastCmd = &cobra.Command{
		Use: "last",
		Run: logsLastCommand,
	}

	logsShowCmd = &cobra.Command{
		Use: "show",
		Run: logsShowCommand,
	}

	logsClearCmd = &cobra.Command{
		Use: "clear",
		Run: logsClearCommand,
	}

	logsBoomtownCmd = &cobra.Command{
		Use: "boomtown",
		Run: logsBoomtownCommand,
	}
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
	by, err := ioutil.ReadFile(filepath.Join(config.LocalLogDir, name))
	if err != nil {
		Exit("Error reading log: %s", name)
	}

	Debug("Reading log: %s", name)
	os.Stdout.Write(by)
}

func logsClearCommand(cmd *cobra.Command, args []string) {
	err := os.RemoveAll(config.LocalLogDir)
	if err != nil {
		Panic(err, "Error clearing %s", config.LocalLogDir)
	}

	Print("Cleared %s", config.LocalLogDir)
}

func logsBoomtownCommand(cmd *cobra.Command, args []string) {
	Debug("Debug message")
	err := errutil.Errorf(errors.New("Inner error message!"), "Error!")
	Panic(err, "Welcome to Boomtown")
	Debug("Never seen")
}

func sortedLogs() []string {
	fileinfos, err := ioutil.ReadDir(config.LocalLogDir)
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
	logsCmd.AddCommand(logsLastCmd, logsShowCmd, logsClearCmd, logsBoomtownCmd)
	RootCmd.AddCommand(logsCmd)
}
