package commands

import (
	"bufio"
	"github.com/github/git-lfs/lfs"
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
)

// untrackCommand takes a list of paths as an argument, and removes each path from the
// default attribtues file (.gitattributes), if it exists.
func untrackCommand(cmd *cobra.Command, args []string) {
	lfs.InstallHooks(false)

	if len(args) < 1 {
		Print("git lfs untrack <path> [path]*")
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

	// Iterate through each line of the attributes file and rewrite it,
	// if the path was meant to be untracked, omit it, and print a message instead.
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "filter=lfs") {
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
				Print("Untracking %s", fields[0])
			}
		}
	}

	attributesFile.Close()
}

func init() {
	RootCmd.AddCommand(untrackCmd)
}
