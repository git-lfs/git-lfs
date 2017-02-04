// +build windows

package commands

import (
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/subprocess"
)

var (
	winBashPrefix string
	winBashMu     sync.Mutex
	winBashRe     *regexp.Regexp
)

func osLineEnding() string {
	return "\r\n"
}

// cleanRootPath replaces the windows root path prefix with a unix path prefix:
// "/". Git Bash (provided with Git For Windows) expands a path like "/foo" to
// the actual Windows directory, but with forward slashes. You can see this
// for yourself:
//
//   $ git /foo
//   git: 'C:/Program Files/Git/foo' is not a git command. See 'git --help'.
//
// You can check the path with `pwd -W`:
//
//   $ cd /
//   $ pwd
//   /
//   $ pwd -W
//   c:/Program Files/Git
func cleanRootPath(pattern string) string {
	winBashMu.Lock()
	defer winBashMu.Unlock()

	// check if path starts with windows drive letter
	if !winPathHasDrive(pattern) {
		return pattern
	}

	if len(winBashPrefix) < 1 {
		// cmd.Path is something like C:\Program Files\Git\usr\bin\pwd.exe
		cmd := subprocess.ExecCommand("pwd")
		winBashPrefix = strings.Replace(filepath.Dir(filepath.Dir(filepath.Dir(cmd.Path))), `\`, "/", -1) + "/"
	}

	return strings.Replace(pattern, winBashPrefix, "/", 1)
}

func winPathHasDrive(pattern string) bool {
	if winBashRe == nil {
		winBashRe = regexp.MustCompile(`\A\w{1}:[/\/]`)
	}

	return winBashRe.MatchString(pattern)
}
