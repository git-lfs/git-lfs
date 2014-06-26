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
	clearLogsFlag bool
	boomtownFlag  bool

	logsCmd = &cobra.Command{
		Use:   "logs",
		Short: "View error logs",
		Run:   logsCommand,
	}
)

func logsCommand(cmd *cobra.Command, args []string) {
	if clearLogsFlag {
		clearLogs()
	}

	if boomtownFlag {
		boomtown()
		return
	}

	var sub string
	if len(args) > 0 {
		sub = args[0]
	}

	switch sub {
	case "last":
		lastLog()
	case "":
		listLogs()
	default:
		showLog(sub)
	}
}

func listLogs() {
	for _, path := range sortedLogs() {
		Print(path)
	}
}

func lastLog() {
	logs := sortedLogs()
	if len(logs) < 1 {
		Print("No logs to show")
		return
	}
	showLog(logs[len(logs)-1])
}

func showLog(name string) {
	by, err := ioutil.ReadFile(filepath.Join(gitmedia.LocalLogDir, name))
	if err != nil {
		Exit("Error reading log: %s", name)
	}

	Debug("Reading log: %s", name)
	os.Stdout.Write(by)
}

func clearLogs() {
	err := os.RemoveAll(gitmedia.LocalLogDir)
	if err != nil {
		Panic(err, "Error clearing %s", gitmedia.LocalLogDir)
	}

	fmt.Println("Cleared", gitmedia.LocalLogDir)
}

func boomtown() {
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
	logsCmd.Flags().BoolVarP(&clearLogsFlag, "clear", "c", false, "Clear existing error logs")
	logsCmd.Flags().BoolVarP(&boomtownFlag, "boomtown", "b", false, "Trigger a panic")
	RootCmd.AddCommand(logsCmd)
}
