package commands

import (
	"errors"
	"fmt"
	"github.com/github/git-media/gitmedia"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
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
	by, err := ioutil.ReadFile(filepath.Join(gitmedia.LocalLogDir, name))
	if err != nil {
		Exit("Error reading log: %s", name)
	}

	Debug("Reading log: %s", name)
	os.Stdout.Write(by)
}

func logsClearCommand(cmd *cobra.Command, args []string) {
	err := os.RemoveAll(gitmedia.LocalLogDir)
	if err != nil {
		Panic(err, "Error clearing %s", gitmedia.LocalLogDir)
	}

	fmt.Println("Cleared", gitmedia.LocalLogDir)
}

func logsBoomtownCommand(cmd *cobra.Command, args []string) {
	Debug("Debug message")
	err := errors.New("Error!")
	Panic(err, "Welcome to Boomtown")
	Debug("Never seen")
}

func sortedLogs() []string {
	fileinfos, err := ioutil.ReadDir(gitmedia.LocalLogDir)
	if err != nil {
		return []string{}
	}

	names := make([]string, len(fileinfos))
	for index, info := range fileinfos {
		names[index] = info.Name()
	}

	return names
}

func init() {
	logsCmd.AddCommand(logsLastCmd, logsShowCmd, logsClearCmd, logsBoomtownCmd)
	RootCmd.AddCommand(logsCmd)
}
