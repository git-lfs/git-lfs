package config

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	GitCommit   string
	Version     = "1.5.5"
	VersionDesc string
)

<<<<<<< HEAD
=======
const (
	Version = "2.0-pre"
)

>>>>>>> f8a50160... Merge branch 'master' into no-dwarf-tables
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
