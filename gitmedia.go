package gitmedia

import (
	"fmt"
	"io/ioutil"
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
	LocalLogDir        string
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

func Environ() []string {
	osEnviron := os.Environ()
	env := make([]string, 2, len(osEnviron)+2)
	env[0] = fmt.Sprintf("LocalMediaDir=%s", LocalMediaDir)
	env[1] = fmt.Sprintf("TempDir=%s", TempDir)

	i := 2
	for _, e := range os.Environ() {
		if !strings.Contains(e, "GIT_") {
			continue
		}
		env[i] = e
		i += 1
	}

	return env
}

func init() {
	LocalMediaDir = resolveMediaDir()
	LocalLogDir = filepath.Join(LocalMediaDir, "logs")
	queueDir = setupQueueDir()

	if err := os.MkdirAll(TempDir, 0744); err != nil {
		Panic(err, "Error trying to create temp directory: %s", TempDir)
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
