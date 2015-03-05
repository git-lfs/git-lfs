package commands

import (
	"bufio"
	"github.com/hawser/git-hawser/hawser"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
)

var (
	untrackCmd = &cobra.Command{
		Use:   "untrack",
		Short: "Remove an entry from .gitattributes",
		Run:   untrackCommand,
	}

	removeCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove an entry from .gitattributes",
		Run:   untrackCommand,
	}

	rmCmd = &cobra.Command{
		Use:   "rm",
		Short: "Remove an entry from .gitattributes",
		Run:   untrackCommand,
	}
)

func untrackCommand(cmd *cobra.Command, args []string) {
	hawser.InstallHooks(false)

	if len(args) < 1 {
		Print("git hawser untrack <path> [path]*")
		return
	}

	data, err := ioutil.ReadFile(".gitattributes")
	if err != nil {
		return
	}

	attributes := strings.NewReader(string(data))

	attributesFile, err := os.Create(".gitattributes")
	if err != nil {
		Print("Error opening .gitattributes for writing")
		return
	}

	scanner := bufio.NewScanner(attributes)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "filter=hawser") {
			fields := strings.Fields(line)
			removeThisPath := false
			for _, t := range args {
				if t == fields[0] {
					removeThisPath = true
				}
			}

			if !removeThisPath {
				attributesFile.WriteString(line + "\n")
			} else {
				Print("Removing path %s", fields[0])
			}
		}
	}

	attributesFile.Close()
}

func init() {
	RootCmd.AddCommand(untrackCmd)
	RootCmd.AddCommand(rmCmd)
	RootCmd.AddCommand(removeCmd)
}
