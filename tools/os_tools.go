package tools

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/git-lfs/git-lfs/v3/subprocess"
	"github.com/git-lfs/git-lfs/v3/tr"
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
			return "", errors.Wrap(err, tr.Tr.Get("error converting working directory to Cygwin"))
		}
	}

	return
}

func translateCygwinPath(path string) (string, error) {
	cmd, err := subprocess.ExecCommand("cygpath", "-w", path)
	if err != nil {
		// If cygpath doesn't exist, that's okay: just return the paths
		// as we got it.
		return path, nil
	}
	// cygpath uses ISO-8850-1 as the default encoding if the locale is not
	// set, resulting in breakage, since we want a UTF-8 path.
	env := make([]string, 0, len(cmd.Env)+1)
	for _, val := range cmd.Env {
		if !strings.HasPrefix(val, "LC_ALL=") {
			env = append(env, val)
		}
	}
	cmd.Env = append(env, "LC_ALL=C.UTF-8")
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
		return path, errors.New(tr.Tr.Get("failed to translate path from Cygwin to Windows: %s", buf.String()))
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
