package tools

import (
	"bytes"
	"fmt"
	"os"
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
		return path, fmt.Errorf("Failed to translate path from cygwin to windows: %s", buf.String())
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
