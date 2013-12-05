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
		Panic(err, "Error running %s %s", name, arg)
	}

	return strings.Trim(string(output), " \n")
}

func LocalMediaPath(sha string) string {
	path := filepath.Join(LocalMediaDir, sha[0:2], sha[2:4])
	if err := os.MkdirAll(path, 0744); err != nil {
		Panic(err, "Error trying to create local media directory: %s", path)
	}

	return filepath.Join(path, sha)
}

func Panic(err error, format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	Debug(line)
	fmt.Fprintln(os.Stderr, line)

	if err != nil {
		Debug(err.Error())
		panic(err)
	}
}

func Debug(format string, args ...interface{}) {
	if !Debugging {
		return
	}
	log.Printf(format, args...)
}

func init() {
	log.SetOutput(os.Stderr)

	LocalMediaDir = resolveMediaDir()
	queueDir = setupQueueDir()

	if err := os.MkdirAll(TempDir, 0744); err != nil {
		Panic(err, "Error trying to create temp directory: %s", TempDir)
	}
}

var Debugging = false

func SetupDebugging(flagset *flag.FlagSet) {
	if flagset == nil {
		flag.BoolVar(&Debugging, "debug", false, "Turns debugging on")
	} else {
		flagset.BoolVar(&Debugging, "debug", false, "Turns debugging on")
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
			Panic(err, "Error reading working directory")
		}
		dir = wd
	}

	return filepath.Join(resolveGitDir(dir), "media")
}

func resolveGitDir(dir string) string {
	base := filepath.Base(dir)
	gitext := ".git"
	if base == gitext || filepath.Ext(base) == gitext {
		return dir
	}
	return filepath.Join(dir, gitext)
}
