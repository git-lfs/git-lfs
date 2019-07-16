package commands

import "strings"

func gitLineEnding(git env) string {
	value, _ := git.Get("core.autocrlf")
	switch strings.ToLower(value) {
	case "true", "t", "1":
		return "\r\n"
	default:
		return osLineEnding()
	}
}

const (
	windowsPrefix = `.\`
	nixPrefix     = `./`
)

// trimCurrentPrefix removes a leading prefix of "./" or ".\" (referring to the
// current directory in a platform independent manner).
//
// It is useful for callers such as "git lfs track" and "git lfs untrack", that
// wish to compare filepaths and/or attributes patterns without cleaning across
// multiple platforms.
func trimCurrentPrefix(p string) string {
	if strings.HasPrefix(p, windowsPrefix) {
		return strings.TrimPrefix(p, windowsPrefix)
	}
	return strings.TrimPrefix(p, nixPrefix)
}

type env interface {
	Get(string) (string, bool)
}
