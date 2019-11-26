package tools

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/git-lfs/git-lfs/subprocess"
	"github.com/pkg/errors"
)

func Getwd() (dir string, err error) {
	dir, err = os.Getwd()
	if err != nil {
		return
	}

	if isCygwin() {
		dir, err = translateCygwinPath(dir)
		if err != nil {
			return "", errors.Wrap(err, "convert wd to cygwin")
		}
	}

	return
}

func translateCygwinPath(path string) (string, error) {
	cmd := subprocess.ExecCommand("cygpath", "-w", path)
	buf := &bytes.Buffer{}
	cmd.Stderr = buf
	out, err := cmd.Output()
	output := strings.TrimSpace(string(out))
	if err != nil {
		// If cygpath doesn't exist, that's okay: just return the paths
		// as we got it.
		if _, ok := err.(*exec.Error); ok {
			return path, nil
		}
		return path, fmt.Errorf("failed to translate path from cygwin to windows: %s", buf.String())
	}
	return output, nil
}

func TranslateCygwinPath(path string) (string, error) {
	if isCygwin() {
		var err error

		path, err = translateCygwinPath(path)
		if err != nil {
			return "", err
		}
	}
	return path, nil
}
