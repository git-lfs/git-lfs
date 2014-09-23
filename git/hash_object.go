package git

import (
	"bytes"
)

func NewHashObject(data []byte) (string, error) {
	buf := bytes.NewBuffer(data)
	return simpleExec(buf, "git", "hash-object", "--stdin")
}
