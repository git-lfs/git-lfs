package commands

import (
	"bufio"
	"fmt"
	"github.com/github/git-media/gitmedia"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type PathCommand struct {
	*Command
}

func (c *PathCommand) Run() {
	gitmedia.InstallHooks()

	var sub string
	if len(c.SubCommands) > 0 {
		sub = c.SubCommands[0]
	}

	switch sub {
	case "add":
		c.addPath()
	case "remove":
		c.removePath()
	default:
		c.listPaths()
	}

}

func (c *PathCommand) addPath() {
	if len(c.SubCommands) < 2 {
		fmt.Println("git media path add <path> [path]*")
		return
	}

	knownPaths := findPaths()
	attributesFile, err := os.OpenFile(".gitattributes", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		fmt.Println("Error opening .gitattributes file")
		return
	}

	for _, t := range c.SubCommands[1:] {
		isKnownPath := false
		for _, k := range knownPaths {
			if t == k.Path {
				isKnownPath = true
			}
		}

		if isKnownPath {
			fmt.Println(t, "already supported")
			continue
		}

		_, err := attributesFile.WriteString(fmt.Sprintf("%s filter=media -crlf\n", t))
		if err != nil {
			fmt.Println("Error adding path", t)
			continue
		}
		fmt.Println("Adding path", t)
	}

	attributesFile.Close()
}

func (c *PathCommand) removePath() {
	if len(c.SubCommands) < 2 {
		fmt.Println("git meda path remove <path> [path]*")
		return
	}

	data, err := ioutil.ReadFile(".gitattributes")
	if err != nil {
		return
	}

	attributes := strings.NewReader(string(data))

	attributesFile, err := os.Create(".gitattributes")
	if err != nil {
		fmt.Println("Error opening .gitattributes for writing")
		return
	}

	scanner := bufio.NewScanner(attributes)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "filter=media") {
			fields := strings.Fields(line)
			removeThisPath := false
			for _, t := range c.SubCommands[1:] {
				if t == fields[0] {
					removeThisPath = true
				}
			}

			if !removeThisPath {
				attributesFile.WriteString(line + "\n")
			} else {
				fmt.Println("Removing path", fields[0])
			}
		}
	}

	attributesFile.Close()
}

func (c *PathCommand) listPaths() {
	fmt.Println("Listing paths")
	knownPaths := findPaths()
	for _, t := range knownPaths {
		fmt.Printf("    %s (%s)\n", t.Path, t.Source)
	}
}

type mediaPath struct {
	Path   string
	Source string
}

func findAttributeFiles() []string {
	paths := make([]string, 0)

	repoAttributes := filepath.Join(gitmedia.LocalGitDir, "info", "attributes")
	if _, err := os.Stat(repoAttributes); err == nil {
		paths = append(paths, repoAttributes)
	}

	filepath.Walk(gitmedia.LocalWorkingDir, func(path string, info os.FileInfo, err error) error {
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
	registerCommand("path", func(c *Command) RunnableCommand {
		return &PathCommand{Command: c}
	})
	registerCommand("paths", func(c *Command) RunnableCommand {
		return &PathCommand{Command: c}
	})
}
