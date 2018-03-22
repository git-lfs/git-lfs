package git

import (
	"bufio"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/subprocess"
)

type Conflict struct {
	Path string

	Common    string
	Head      string
	MergeHead string
}

func AllConflicts() ([]*Conflict, error) {
	scan, err := NewConflictScanner()
	if err != nil {
		return nil, err
	}

	var conflicts []*Conflict

	for scan.Scan() {
		conflicts = append(conflicts, scan.Next())
	}

	if err := scan.Err(); err != nil {
		return nil, err
	}
	return conflicts, nil
}

func ConflictsInDir(dirname string) ([]*Conflict, error) {
	scan, err := NewConflictScanner(dirname)
	if err != nil {
		return nil, err
	}

	var conflicts []*Conflict

	for scan.Scan() {
		conflict := scan.Next()
		if filepath.Dir(conflict.Path) == dirname {
			conflicts = append(conflicts, conflict)
		}
	}

	if err := scan.Err(); err != nil {
		return nil, err
	}
	return conflicts, nil
}

func ConflictDetails(fname string) (*Conflict, error) {
	scan, err := NewConflictScanner(fname)
	if err != nil {
		return nil, err
	}

	if !scan.Scan() {
		return nil, scan.Err()
	}
	return scan.Next(), scan.Err()
}

type ConflictScanner struct {
	cmd  *subprocess.Cmd
	scan *bufio.Scanner

	c   *Conflict
	err error
}

func NewConflictScanner(args ...string) (*ConflictScanner, error) {
	cmd := gitNoLFS(append([]string{
		"-c", "core.quotepath=false",
		"status",
		"--porcelain=2",
		"--"},
		args...)...)

	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to call git-status(1)")
	}
	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "Failed to start git-status(1)")
	}

	bufio.NewScanner(out)

	return &ConflictScanner{
		cmd:  cmd,
		scan: bufio.NewScanner(out),
	}, nil
}

func (c *ConflictScanner) Scan() bool {
	c.c, c.err = nil, nil

	for c.scan.Scan() {
		c.c, c.err = conflict(c.scan.Text())
		if c.c != nil || c.err != nil {
			break
		}
	}
	return c.c != nil
}

func conflict(text string) (*Conflict, error) {
	fields := strings.Split(text, " ")
	if fields[0] != "u" {
		return nil, nil
	}

	if len(fields) < 11 {
		return nil, errors.Errorf("git: malformed status line: %s", text)
	}

	return &Conflict{
		Path: strings.Join(fields[10:], " "),

		Common:    fields[7],
		Head:      fields[8],
		MergeHead: fields[9],
	}, nil
}

func (c *ConflictScanner) Next() *Conflict {
	return c.c
}

func (c *ConflictScanner) Err() error {
	if err := c.cmd.Wait(); err != nil {
		return err
	}
	if err := c.scan.Err(); err != nil {
		return err
	}
	return nil
}
