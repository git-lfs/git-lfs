package gitmedia

import (
	_ "../gitconfig"
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type PathCommand struct {
	*Command
}

func (c *PathCommand) Run() {
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
			if t == k {
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
		fmt.Println("    ", t)
	}
}

func findPaths() []string {
	paths := make([]string, 0)

	attributes, err := os.Open(".gitattributes")
	if err != nil {
		return paths // No .gitattibtues == no file paths
	}

	scanner := bufio.NewScanner(attributes)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if strings.Contains(line, "filter=media") {
			fields := strings.Fields(line)
			paths = append(paths, fields[0])
		}
	}

	return paths
}

func init() {
	registerCommand("path", func(c *Command) RunnableCommand {
		return &PathCommand{Command: c}
	})
}
