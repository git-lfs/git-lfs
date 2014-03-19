package gitconfig

import (
	gitmedia ".."
)

func Find(val string) string {
	return gitmedia.SimpleExec("git", "config", val)
}

func SetGlobal(key, val string) {
	gitmedia.SimpleExec("git", "config", "--global", "--add", key, val)
}
