package commands

import (
	"fmt"
	"github.com/hawser/git-hawser/hawser"
	"github.com/spf13/cobra"
	"os"
)

var (
	addCmd = &cobra.Command{
		Use:   "add",
		Short: "Add an entry to .gitattributes",
		Run:   addCommand,
	}
)

func addCommand(cmd *cobra.Command, args []string) {
	hawser.InstallHooks(false)

	if len(args) < 1 {
		Print("git hawser path add <path> [path]*")
		return
	}

	knownPaths := findPaths()
	attributesFile, err := os.OpenFile(".gitattributes", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		Print("Error opening .gitattributes file")
		return
	}

	for _, t := range args {
		isKnownPath := false
		for _, k := range knownPaths {
			if t == k.Path {
				isKnownPath = true
			}
		}

		if isKnownPath {
			Print("%s already supported", t)
			continue
		}

		_, err := attributesFile.WriteString(fmt.Sprintf("%s filter=hawser -crlf\n", t))
		if err != nil {
			Print("Error adding path %s", t)
			continue
		}
		Print("Adding path %s", t)
	}

	attributesFile.Close()
}

func init() {
	RootCmd.AddCommand(addCmd)
}
