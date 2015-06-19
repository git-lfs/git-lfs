package commands

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	logsCmd = &cobra.Command{
		Use:   "logs",
		Short: "View error logs",
		Run:   logsCommand,
	}

	logsLastCmd = &cobra.Command{
		Use:   "last",
		Short: "View latest error log",
		Run:   logsLastCommand,
	}

	logsShowCmd = &cobra.Command{
		Use:   "show",
		Short: "View a single error log",
		Run:   logsShowCommand,
	}

	logsClearCmd = &cobra.Command{
		Use:   "clear",
		Short: "Clear all logs",
		Run:   logsClearCommand,
	}

	logsBoomtownCmd = &cobra.Command{
		Use:   "boomtown",
		Short: "Trigger a sample error",
		Run:   logsBoomtownCommand,
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
	by, err := ioutil.ReadFile(filepath.Join(lfs.LocalLogDir, name))
	if err != nil {
		Exit("Error reading log: %s", name)
	}

	Debug("Reading log: %s", name)
	os.Stdout.Write(by)
}

func logsClearCommand(cmd *cobra.Command, args []string) {
	err := os.RemoveAll(lfs.LocalLogDir)
	if err != nil {
		Panic(err, "Error clearing %s", lfs.LocalLogDir)
	}

	Print("Cleared %s", lfs.LocalLogDir)
}

func logsBoomtownCommand(cmd *cobra.Command, args []string) {
	Debug("Debug message")
	err := lfs.Errorf(errors.New("Inner error message!"), "Error!")
	Panic(err, "Welcome to Boomtown")
	Debug("Never seen")
}

func sortedLogs() []string {
	fileinfos, err := ioutil.ReadDir(lfs.LocalLogDir)
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
