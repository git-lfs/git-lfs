package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	bashPath    string
	debugging   = false
	erroring    = false
	maxprocs    = 4
	testPattern = regexp.MustCompile(`test[/\\]test-([a-z\-]+)\.sh$`)
)

func mainIntegration() {
	if len(os.Getenv("DEBUG")) > 0 {
		debugging = true
	}

	setBash()

	if max, _ := strconv.Atoi(os.Getenv("GIT_LFS_TEST_MAXPROCS")); max > 0 {
		maxprocs = max
	}

	fmt.Println("Running this maxprocs", maxprocs)

	files := testFiles()

	if len(files) == 0 {
		fmt.Println("no tests to run")
		os.Exit(1)
	}

	var wg sync.WaitGroup
	tests := make(chan string, len(files))
	output := make(chan string, len(files))

	for _, file := range files {
		tests <- file
	}
	close(tests)

	outputDone := make(chan bool)
	go func() {
		for out := range output {
			fmt.Println(out)
		}
		outputDone <- true
	}()

	for i := 0; i < maxprocs; i++ {
		wg.Add(1)
		go worker(tests, output, &wg)
	}

	wg.Wait()
	close(output)
	<-outputDone

	if erroring {
		os.Exit(1)
	}
}

func runTest(output chan string, testname string) {
	buf := &bytes.Buffer{}
	cmd := exec.Command(bashPath, testname)
	cmd.Stdout = buf
	cmd.Stderr = buf

	err := cmd.Start()
	if err != nil {
		sendTestOutput(output, testname, buf, err)
		return
	}

	done := make(chan error)
	go func() {
		if err := cmd.Wait(); err != nil {
			done <- err
		}
		close(done)
	}()

	select {
	case err = <-done:
		sendTestOutput(output, testname, buf, err)
		return
	case <-time.After(3 * time.Minute):
		sendTestOutput(output, testname, buf, errors.New("Timed out"))
		cmd.Process.Kill()
		return
	}
}

func sendTestOutput(output chan string, testname string, buf *bytes.Buffer, err error) {
	cli := strings.TrimSpace(buf.String())
	if len(cli) == 0 {
		cli = fmt.Sprintf("<no output for %s>", testname)
	}

	if err == nil {
		output <- cli
	} else {
		basetestname := filepath.Base(testname)
		if debugging {
			fmt.Printf("Error on %s: %s\n", basetestname, err)
		}
		erroring = true
		output <- fmt.Sprintf("error: %s => %s\n%s", basetestname, err, cli)
	}
}

func worker(tests <-chan string, output chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case testname, ok := <-tests:
			if !ok {
				return
			}
			runTest(output, testname)
		}
	}
}

func testFiles() []string {
	if len(os.Args) < 4 {
		return allTestFiles()
	}

	fileMap := make(map[string]bool)
	for _, file := range allTestFiles() {
		fileMap[file] = true
	}

	files := make([]string, 0, len(os.Args)-3)
	for _, arg := range os.Args {
		fullname := "test/test-" + arg + ".sh"
		if fileMap[fullname] {
			files = append(files, fullname)
		}
	}

	return files
}

func allTestFiles() []string {
	files := make([]string, 0, 100)
	filepath.Walk("test", func(path string, info os.FileInfo, err error) error {
		if debugging {
			fmt.Println("FOUND:", path)
		}
		if err != nil || info.IsDir() || !testPattern.MatchString(path) {
			return nil
		}

		if debugging {
			fmt.Println("MATCHING:", path)
		}
		files = append(files, path)
		return nil
	})
	return files
}

func setBash() {
	findcmd := "which"
	if runtime.GOOS == "windows" {
		// Can't use paths returned from which even if it's on PATH in Windows
		// Because our Go binary is a separate Windows app & not MinGW, it
		// can't understand paths like '/usr/bin/bash', needs Windows version
		findcmd = "where"
	}

	out, err := exec.Command(findcmd, "bash").Output()
	if err != nil {
		fmt.Println("Unable to find bash:", err)
		os.Exit(1)
	}
	if len(out) == 0 {
		fmt.Printf("No output from '%s bash'\n", findcmd)
		os.Exit(1)
	}

	bashPath = strings.TrimSpace(strings.Split(string(out), "\n")[0])
	if debugging {
		fmt.Println("Using", bashPath)
	}

	// Test
	_, err = exec.Command(bashPath, "--version").CombinedOutput()
	if err != nil {
		fmt.Println("Error calling bash:", err)
		os.Exit(1)
	}
}
