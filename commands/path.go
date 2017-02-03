package commands

import "strings"

const (
	lf   = "\n"
	crlf = "\r\n"
)

func gitLineEnding(git env) string {
	value, _ := git.Get("core.autocrlf")
	switch strings.ToLower(value) {
	case "input", "true", "t", "1":
		return crlf
	default:
		return lineEnding()
	}
}

type env interface {
	Get(string) (string, bool)
}
