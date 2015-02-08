package tests

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type cliSuite struct {
	Name     string
	Filename string
	TempDir  string
	Tests    []*cliTest
	Output   *bytes.Buffer
	Err      error
}

type cliTest struct {
	Command  string
	Expected string
}

func NewSuite(now time.Time, path string) *cliSuite {
	s := &cliSuite{
		Name:     path[0 : len(path)-4],
		Filename: path,
		Output:   &bytes.Buffer{},
	}

	if err := load(s); err != nil {
		s.Errorf("Error loading test suite for %s: %s", s.Filename, err)
		return s
	}

	testDir := filepath.Join(os.TempDir(), "hawser-tests")
	os.RemoveAll(testDir)
	s.TempDir = filepath.Join(testDir, fmt.Sprintf("%d-%s", now.Unix(), s.Name))

	if s.Name == "empty" {
		s.Logf("# Initializing empty repository")
		if err := os.MkdirAll(s.TempDir, 0744); err != nil {
			s.Errorf("error creating %s: %s", s.TempDir, err)
			return s
		}

		os.Chdir(s.TempDir)

		s.Exec("git", "init")
	} else {
		s.Errorf("unable to test")
	}

	return s
}

func (s *cliSuite) Run() {
	for _, test := range s.Tests {
		if s.Err != nil {
			continue
		}
		test.Run(s)
	}

	if s.Err != nil {
		s.Logf("ERROR: %s", s.Err)
	}
}

func (s *cliSuite) Exec(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	outputBytes, err := cmd.CombinedOutput()
	s.Output.Write(outputBytes)
	if err != nil {
		s.Errorf("Error running: %s %s:\n%s", name, args, err)
	}

	return strings.TrimSpace(string(outputBytes))
}

func (s *cliSuite) Logf(format string, args ...interface{}) {
	fmt.Fprintf(s.Output, format+"\n", args...)
}

func (s *cliSuite) Errorf(format string, args ...interface{}) {
	s.Err = fmt.Errorf(format, args...)
}

func (t *cliTest) Run(s *cliSuite) {
	s.Logf("$ %s", t.Command)
	cmd := t.Command
	if strings.HasPrefix(t.Command, "git hawser ") {
		cmd = Bin + " " + t.Command[11:]
	}

	parts := strings.Split(cmd, " ")
	output := s.Exec(parts[0], parts[1:]...)

	if output != t.Expected {
		s.Errorf("expected this instead:\n%s", t.Expected)
	} else {
		s.Logf("")
	}
}

// Opens a test suite file, and parses the individual commands and expected
// outputs.
func load(s *cliSuite) error {
	s.Tests = make([]*cliTest, 0, 100)

	file, err := os.Open(s.Filename)
	if err != nil {
		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var test *cliTest
	expected := make([]string, 1000)
	expectedPos := 0

	addTest := func(test *cliTest) {
		if test == nil {
			return
		}
		test.Expected = strings.TrimSpace(strings.Join(expected[0:expectedPos], "\n"))
		s.Tests = append(s.Tests, test)
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "$") {
			addTest(test)
			test = &cliTest{
				Command: strings.TrimSpace(line[1:]),
			}
			expectedPos = 0
			continue
		}

		if test != nil {
			expected[expectedPos] = line
			expectedPos += 1
		}
	}

	addTest(test)

	return scanner.Err()
}
