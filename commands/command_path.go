package commands

import (
	"bufio"
	"fmt"
	"github.com/hawser/git-hawser/hawser"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	pathCmd = &cobra.Command{
		Use:   "path",
		Short: "Manipulate .gitattributes",
		Run:   pathCommand,
	}

	pathAddCmd = &cobra.Command{
		Use:   "add",
		Short: "Add an entry to .gitattributes",
		Run:   pathAddCommand,
	}

	pathRemoveCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove an entry from .gitattributes",
		Run:   pathRemoveCommand,
	}
)

func pathCommand(cmd *cobra.Command, args []string) {
	hawser.InstallHooks(false)

	Print("Listing paths")
	knownPaths := findPaths()
	for _, t := range knownPaths {
		Print("    %s (%s)", t.Path, t.Source)
	}
}

func pathAddCommand(cmd *cobra.Command, args []string) {
	hawser.InstallHooks(false)

	if len(args) < 1 {
		Print("git media path add <path> [path]*")
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

		_, err := attributesFile.WriteString(fmt.Sprintf("%s filter=media -crlf\n", t))
		if err != nil {
			Print("Error adding path %s", t)
			continue
		}
		Print("Adding path %s", t)
	}

	attributesFile.Close()
}

func pathRemoveCommand(cmd *cobra.Command, args []string) {
	hawser.InstallHooks(false)

	if len(args) < 1 {
		Print("git media path remove <path> [path]*")
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
		if strings.Contains(line, "filter=media") {
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

type mediaPath struct {
	Path   string
	Source string
}

func findAttributeFiles() []string {
	paths := make([]string, 0)

	repoAttributes := filepath.Join(hawser.LocalGitDir, "info", "attributes")
	if _, err := os.Stat(repoAttributes); err == nil {
		paths = append(paths, repoAttributes)
	}

	filepath.Walk(hawser.LocalWorkingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (filepath.Base(path) == ".gitattributes") {
			paths = append(paths, path)
		}
		return nil
	})

	return paths
}

func findPaths() []mediaPath {
	paths := make([]mediaPath, 0)
	wd, _ := os.Getwd()

	for _, path := range findAttributeFiles() {
		attributes, err := os.Open(path)
		if err != nil {
			return paths
		}

		scanner := bufio.NewScanner(attributes)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			if strings.Contains(line, "filter=media") {
				fields := strings.Fields(line)
				relPath, _ := filepath.Rel(wd, path)
				paths = append(paths, mediaPath{Path: fields[0], Source: relPath})
			}
		}
	}

	return paths
}

func init() {
	pathCmd.AddCommand(pathAddCmd, pathRemoveCmd)
	RootCmd.AddCommand(pathCmd)
}
