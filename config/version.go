package config

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	GitCommit   string
<<<<<<< HEAD
	Version     = "1.5.3"
=======
>>>>>>> 2b606199... Merge pull request #1689 from sschuberth/version-defines
	VersionDesc string
)

const (
	Version = "1.5.0"
)

func init() {
	gitCommit := ""
	if len(GitCommit) > 0 {
		gitCommit = "; git " + GitCommit
	}
	VersionDesc = fmt.Sprintf("git-lfs/%s (GitHub; %s %s; go %s%s)",
		Version,
		runtime.GOOS,
		runtime.GOARCH,
		strings.Replace(runtime.Version(), "go", "", 1),
		gitCommit,
	)

}
