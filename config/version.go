package config

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	GitCommit   string
	VersionDesc string
	Vendor      string
)

const (
	Version = "2.13.1"
)

func init() {
	gitCommit := ""
	if len(GitCommit) > 0 {
		gitCommit = "; git " + GitCommit
	}
	if len(Vendor) == 0 {
		Vendor = "GitHub"
	}
	VersionDesc = fmt.Sprintf("git-lfs/%s (%s; %s %s; go %s%s)",
		Version,
		Vendor,
		runtime.GOOS,
		runtime.GOARCH,
		strings.Replace(runtime.Version(), "go", "", 1),
		gitCommit,
	)
}
