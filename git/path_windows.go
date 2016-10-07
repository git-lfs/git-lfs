// +build windows

package git

import (
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/github/git-lfs/subprocess"
)

var (
	defaultPathPrefix = "/"
	pathPrefix        string
	pathPrefixMu      sync.Mutex
	pathPrefixRe      *regexp.Regexp
)

func sanitizePatternPrefix(pattern string) string {
	pathPrefixMu.Lock()
	defer pathPrefixMu.Unlock()

	if pathPrefixRe == nil {
		pathPrefixRe = regexp.MustCompile(`\A\w{1}:[/\/]`)
	}

	// check if path starts with c:/ or c:\
	if !pathPrefixRe.MatchString(pattern) {
		return defaultPathPrefix
	}

	if len(pathPrefix) < 1 {
		cmd := subprocess.ExecCommand("pwd")
		pathPrefix = strings.Replace(filepath.Dir(filepath.Dir(filepath.Dir(cmd.Path))), `\`, "/", -1) + "/"
	}

	return pathPrefix
}
