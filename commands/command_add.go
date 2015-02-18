package commands

import (
	"fmt"
	"github.com/hawser/git-hawser/hawser"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
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

	addTrailingLinebreak := needsTrailingLinebreak(".gitattributes")
	knownPaths := findPaths()
	attributesFile, err := os.OpenFile(".gitattributes", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		Print("Error opening .gitattributes file")
		return
	}

	if addTrailingLinebreak {
		if _, err := attributesFile.WriteString("\n"); err != nil {
			Print("Error writing to .gitattributes")
		}
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

func needsTrailingLinebreak(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}

	defer file.Close()
	buf := make([]byte, 16384)
	bytesRead := 0
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			return false
		}
		bytesRead = n
	}

	return !strings.HasSuffix(string(buf[0:bytesRead]), "\n")
}

func init() {
	RootCmd.AddCommand(addCmd)
}
