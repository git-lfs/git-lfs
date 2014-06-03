package gitmedia

import (
	"errors"
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
	LocalWorkingDir    string
	LocalGitDir        string
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
	env := make([]string, 4, len(osEnviron)+4)
	env[0] = fmt.Sprintf("LocalWorkingDir=%s", LocalWorkingDir)
	env[1] = fmt.Sprintf("LocalGitDir=%s", LocalGitDir)
	env[2] = fmt.Sprintf("LocalMediaDir=%s", LocalMediaDir)
	env[3] = fmt.Sprintf("TempDir=%s", TempDir)

	for _, e := range osEnviron {
		if !strings.Contains(e, "GIT_") {
			continue
		}
		env = append(env, e)
	}

	return env
}

func InRepo() bool {
	return LocalWorkingDir != ""
}

func InstallHooks() error {
	if !InRepo() {
		return errors.New("Not in a repository")
	}

	return nil
}

func init() {
	var err error
	LocalWorkingDir, LocalGitDir, err = resolveGitDir()
	if err == nil {
		LocalMediaDir = filepath.Join(LocalGitDir, "media")
		LocalLogDir = filepath.Join(LocalMediaDir, "logs")
		queueDir = setupQueueDir()

		if err := os.MkdirAll(TempDir, 0744); err != nil {
			Panic(err, "Error trying to create temp directory: %s", TempDir)
		}
	}
}

func resolveGitDir() (string, string, error) {
	wd, err := os.Getwd()
	if err != nil {
		Panic(err, "Error reading working directory")
	}

	return recursiveResolveGitDir(wd)
}

func recursiveResolveGitDir(dir string) (string, string, error) {
	if len(dir) == 1 && dir[0] == os.PathSeparator {
		return "", "", fmt.Errorf("Git repository not found")
	}

	if filepath.Base(dir) == gitExt {
		return filepath.Dir(dir), dir, nil
	}

	gitDir := filepath.Join(dir, gitExt)
	if info, err := os.Stat(gitDir); err == nil {
		if info.IsDir() {
			return dir, gitDir, nil
		}
	}

	return recursiveResolveGitDir(filepath.Dir(dir))
}

const gitExt = ".git"
