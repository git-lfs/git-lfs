package gitconfig

import (
	"github.com/github/git-media/gitmedia"
)

func Find(val string) string {
	return gitmedia.SimpleExec("git", "config", val)
}

func SetGlobal(key, val string) {
	gitmedia.SimpleExec("git", "config", "--global", "--add", key, val)
}

func UnsetGlobal(key string) {
	gitmedia.SimpleExec("git", "config", "--global", "--unset", key)
}
