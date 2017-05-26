package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	Bin    string
	suites []*cliSuite
)

func TestTests(t *testing.T) {
	paths, err := filepath.Glob("*.txt")
	if err != nil {
		t.Fatalf("Error getting test files: %s", err)
	}

	now := time.Now()
	for _, path := range paths {
		suite := NewSuite(now, path)
		suite.Run()
		if suite.Err != nil {
			t.Errorf(suite.Output.String())
		}
	}
}

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	Bin = filepath.Join(wd, "..", "bin", "git-hawser")
}
