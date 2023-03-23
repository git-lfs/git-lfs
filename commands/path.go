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

type env interface {
	Get(string) (string, bool)
}
