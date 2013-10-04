package gitmedia

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const Version = "0.0.1"

var (
	LargeSizeThreshold = 5 * 1024 * 1024
	TempDir            = filepath.Join(os.TempDir(), "git-media")
	LocalMediaDir      string
)

func TempFile() (*os.File, error) {
	return ioutil.TempFile(TempDir, "")
}

func SimpleExec(name string, arg ...string) string {
	output, err := exec.Command(name, arg...).Output()
	if _, ok := err.(*exec.ExitError); ok {
		return ""
	} else if err != nil {
		fmt.Printf("error running: %s %s\n", name, arg)
		panic(err)
	}

	return strings.Trim(string(output), " \n")
}

func LocalMediaPath(sha string) string {
	path := filepath.Join(LocalMediaDir, sha[0:2], sha[2:4])
	if err := os.MkdirAll(path, 0744); err != nil {
		fmt.Printf("Error trying to create local media directory: %s\n", path)
		panic(err)
	}

	return filepath.Join(path, sha)
}

func Panic(err error, format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	Debug(line)
	fmt.Println(line)

	if err != nil {
		Debug(err.Error())
		panic(err)
	}
}

func Debug(format string, args ...interface{}) {
	if !*debugging {
		return
	}
	log.Printf(format, args...)
}

func init() {
	LocalMediaDir = resolveMediaDir()

	if err := os.MkdirAll(TempDir, 0744); err != nil {
		fmt.Printf("Error trying to create temp directory: %s\n", TempDir)
		panic(err)
	}
}

var debugging = flag.Bool("debug", false, "Turns debugging on")

func SetupDebugging() {
	flag.Parse()

	if *debugging {
		log.SetOutput(os.Stderr)
	}
}

func resolveMediaDir() string {
	dir := os.Getenv("GIT_MEDIA_DIR")
	if len(dir) > 0 {
		return dir
	}

	dir = os.Getenv("GIT_DIR")
	if len(dir) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		dir = wd
	}

	return resolveGitDir(dir)
}

func resolveGitDir(dir string) string {
	return filepath.Join(dir, ".git", "media")
}
