package git

import (
	"bufio"
	"io"
	"path"
	"strings"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/rubyist/tracerx"
)

type lsFileInfo struct {
	BaseName string
	FullPath string
}
type LsFiles struct {
	Files       map[string]*lsFileInfo
	FilesByName map[string][]*lsFileInfo
}

func NewLsFiles(workingDir string, standardExclude bool, untracked bool) (*LsFiles, error) {

	args := []string{
		"ls-files",
		"-z", // Use a NUL separator. This also disables the escaping of special characters.
		"--cached",
	}

	if IsGitVersionAtLeast("2.35.0") {
		args = append(args, "--sparse")
	}

	if standardExclude {
		args = append(args, "--exclude-standard")
	}
	if untracked {
		args = append(args, "--others")
	}
	cmd, err := gitNoLFS(args...)
	if err != nil {
		return nil, err
	}
	cmd.Dir = workingDir

	tracerx.Printf("NewLsFiles: running in %s git %s",
		workingDir, strings.Join(args, " "))

	// Capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Split(tools.SplitOnNul)

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	rv := &LsFiles{
		Files:       make(map[string]*lsFileInfo),
		FilesByName: make(map[string][]*lsFileInfo),
	}

	// Setup a goroutine to drain stderr as large amounts of error output may cause
	// the subprocess to block.
	errorMessages := make(chan []byte)
	go func() {
		msg, _ := io.ReadAll(stderr)
		errorMessages <- msg
	}()

	// Read all files
	for scanner.Scan() {
		base := path.Base(scanner.Text())
		finfo := &lsFileInfo{
			BaseName: base,
			FullPath: scanner.Text(),
		}
		rv.Files[scanner.Text()] = finfo
		rv.FilesByName[base] = append(rv.FilesByName[base], finfo)
	}

	// Check the output of the subprocess, output stderr if the command failed.
	msg := <-errorMessages
	if err := cmd.Wait(); err != nil {
		return nil, errors.New(tr.Tr.Get("Error in `git %s`: %v %s",
			strings.Join(args, " "), err, msg))
	}

	return rv, nil
}
