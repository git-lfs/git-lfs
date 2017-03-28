package config

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/git-lfs/git-lfs/lfsapi"
)

var (
	GitCommit   string
	VersionDesc string
)

const (
<<<<<<< HEAD
	Version = "2.0.1"
=======
	Version = "2.1.0-pre"
>>>>>>> f8e71189... Merge branch 'master' into auth-caching
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

	lfsapi.UserAgent = VersionDesc
}
