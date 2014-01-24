package main

import (
	gitmedia ".."
	"fmt"
	"os"
	"path/filepath"
)

var (
	allCommands map[string]map[string]string
	gitMediaBin string
)

func main() {
	for wd, commands := range allCommands {
		fmt.Println("Integration tests for", wd)
		for cmd, expected := range commands {
			if err := os.Chdir(wd); err != nil {
				fmt.Println("Cannot chdir to", wd)
				os.Exit(1)
			}
			fmt.Println("$ git-media", cmd)
			actual := gitmedia.SimpleExec(gitMediaBin, cmd)
			if actual != expected {
				fmt.Printf("expected:\n%s\n\n", expected)
				fmt.Printf("actual:\n%s\n", actual)
			}
		}
	}
}

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	gitMediaBin = filepath.Join(wd, "bin", "git-media")

	allCommands = make(map[string]map[string]string)

	// tests on the git-media repository, which has no actual git-media assets :)
	allCommands[wd] = map[string]string{
		"version": "git-media v" + gitmedia.Version,
		"config": "Endpoint=\n" +
			"LocalMediaDir=" + filepath.Join(wd, ".git", "media") + "\n" +
			"TempDir=" + filepath.Join(os.TempDir(), "git-media"),
	}
}
