package commands

import "strings"

func gitLineEnding(git env) string {
	value, _ := git.Get("core.autocrlf")
	switch strings.ToLower(value) {
	case "input", "true", "t", "1":
		return "\r\n"
	default:
		return osLineEnding()
	}
}

const (
	windowsPrefix = `.\`
	nixPrefix     = `./`
)

func trimCurrentPrefix(p string) string {
	if strings.HasPrefix(p, windowsPrefix) {
		return strings.TrimPrefix(p, windowsPrefix)
	}
	return strings.TrimPrefix(p, nixPrefix)
}

type env interface {
	Get(string) (string, bool)
}
