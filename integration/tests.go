package main

import (
	gitmedia ".."
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	allCommands map[string]map[string]string
	gitMediaBin string
)

func main() {
	exitCode := 0

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
				exitCode = 1
				fmt.Printf("- expected\n%s\n\n", expected)
				fmt.Printf("- actual\n%s\n", actual)
			}
		}
		fmt.Println("")
	}

	os.Exit(exitCode)
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
		"config": "Endpoint=https://github.com/github/git-media.git/info/media\n" +
			"LocalWorkingDir=" + wd + "\n" +
			"LocalGitDir=" + filepath.Join(wd, ".git") + "\n" +
			"LocalMediaDir=" + filepath.Join(wd, ".git", "media") + "\n" +
			"TempDir=" + filepath.Join(os.TempDir(), "git-media"),
	}

	// tests on the git-media .git dir
	allCommands[filepath.Join(wd, ".git")] = allCommands[wd]

	// tests on the git-media sub directory
	allCommands[filepath.Join(wd, "integration")] = allCommands[wd]

	if gitEnv := gitEnviron(); len(gitEnv) > 0 {
		suffix := "\n" + strings.Join(gitEnv, "\n")
		allCommands[wd]["config"] += suffix
	}
}

func gitEnviron() []string {
	osEnviron := os.Environ()
	env := make([]string, 0, len(osEnviron))

	for _, e := range osEnviron {
		if !strings.Contains(e, "GIT_") {
			continue
		}
		env = append(env, e)
	}

	return env
}
