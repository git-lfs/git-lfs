package commands

import (
	"bufio"
	"io/ioutil"
	"os"
	"strings"

	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/spf13/cobra"
)

// untrackCommand takes a list of paths as an argument, and removes each path from the
// default attributes file (.gitattributes), if it exists.
func untrackCommand(cmd *cobra.Command, args []string) {
	setupWorkingCopy()

	installHooks(false, false)

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
		Print(tr.Tr.Get("Error opening '.gitattributes' for writing"))
		return
	}
	defer attributesFile.Close()

	scanner := bufio.NewScanner(attributes)

	// Iterate through each line of the attributes file and rewrite it,
	// if the path was meant to be untracked, omit it, and print a message instead.
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "filter=lfs") {
			attributesFile.WriteString(line + "\n")
			continue
		}

		path := strings.Fields(line)[0]
		if removePath(path, args) {
			Print(tr.Tr.Get("Untracking %q", unescapeAttrPattern(path)))
		} else {
			attributesFile.WriteString(line + "\n")
		}
	}
}

func removePath(path string, args []string) bool {
	withoutCurrentDir := trimCurrentPrefix(path)
	for _, t := range args {
		if withoutCurrentDir == escapeAttrPattern(trimCurrentPrefix(t)) {
			return true
		}
	}

	return false
}

func init() {
	RegisterCommand("untrack", untrackCommand, nil)
}
