// +build testtools

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/git-lfs/git-lfs/test"
)

type TestUtilRepoCallback struct{}

func (*TestUtilRepoCallback) Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(4)
}
func (*TestUtilRepoCallback) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func main() {
	commandMap := map[string]func(*test.Repo){
		"addcommits": AddCommits,
	}
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Command required (e.g. addcommits)\n")
		os.Exit(2)
	}

	f, ok := commandMap[os.Args[1]]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown command: %v\n", os.Args[1])
		os.Exit(2)
	}
	// Construct test repo context (note: no Cleanup() call since managed outside)
	// also assume we're in the same folder
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Problem getting working dir: %v\n", err)
		os.Exit(2)
	}
	// Make sure we're directly inside directory which contains .git
	// don't want to accidentally end up committing to some other parent git
	_, err = os.Stat(filepath.Join(wd, ".git"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "You're in the wrong directory, should be in root of a test repo: %v\n", err)
		os.Exit(2)
	}

	repo := test.WrapRepo(&TestUtilRepoCallback{}, wd)
	f(repo)
}

func AddCommits(repo *test.Repo) {
	// Read stdin as JSON []*test.CommitInput
	in, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "addcommits: Unable to read input data: %v\n", err)
		os.Exit(3)
	}
	inputs := make([]*test.CommitInput, 0)
	err = json.Unmarshal(in, &inputs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "addcommits: Unable to unmarshal JSON: %v\n%v\n", string(in), err)
		os.Exit(3)
	}
	outputs := repo.AddCommits(inputs)

	by, err := json.Marshal(outputs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "addcommits: Unable to marshal output JSON: %v\n", err)
		os.Exit(3)
	}
	// Write response to stdout
	_, err = os.Stdout.Write(by)
	if err != nil {
		fmt.Fprintf(os.Stderr, "addcommits: Error writing JSON to stdout: %v\n", err)
		os.Exit(3)
	}
	os.Stdout.WriteString("\n")

}
