package gitmedia

import (
	_ "../gitconfig"
	"bufio"
	"fmt"
	"os"
	"strings"
)

type TypesCommand struct {
	*Command
}

func (c *TypesCommand) Run() {
	var sub string
	if len(c.SubCommands) > 0 {
		sub = c.SubCommands[0]
	}

	switch sub {
	case "add":
		fmt.Println("Adding type")
	case "remove":
		fmt.Println("Removing type")
	default:
		fmt.Println("Listing types")
		listTypes()
	}

}

func listTypes() {
	attributes, err := os.Open(".gitattributes")
	if err != nil {
		return // No .gitattibtues == no file types
	}

	scanner := bufio.NewScanner(attributes)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if strings.Contains(line, "filter=media") {
			fields := strings.Fields(line)
			fmt.Println("   ", fields[0])
		}
	}

}

func init() {
	registerCommand("types", func(c *Command) RunnableCommand {
		return &TypesCommand{Command: c}
	})
}
