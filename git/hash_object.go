package git

import (
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
)

type HashObject struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	hash   string
}

func NewHashObject() (*HashObject, error) {
	cmd := exec.Command("git", "hash-object", "--stdin")
	gitHashWriter, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	gitHashReader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	cmd.Start()

	return &HashObject{cmd, gitHashWriter, gitHashReader, ""}, nil
}

func (gh *HashObject) Write(p []byte) (int, error) {
	return gh.stdin.Write(p)
}

func (gh *HashObject) Close() error {
	err := gh.stdin.Close()
	if err == nil {
		hashBytes, err := ioutil.ReadAll(gh.stdout)
		if err == nil {
			gh.hash = strings.TrimSpace(string(hashBytes))
		}
		gh.cmd.Wait()
	}
	return err
}

func (gh *HashObject) Hash() string {
	return gh.hash
}
