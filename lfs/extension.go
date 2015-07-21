package lfs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// An Extension describes how to manipulate files during smudge and clean.
type Extension struct {
	Name     string
	Clean    string
	Smudge   string
	Priority int
}

type extCommand struct {
	ext    *PointerExtension
	cmd    *exec.Cmd
	out    io.WriteCloser
	err    *bytes.Buffer
	hasher hash.Hash
}

// SortExtensions sorts a map of extensions in ascending order by Priority
func SortExtensions(m map[string]Extension) ([]Extension, error) {
	pMap := make(map[int]Extension)
	for n, ext := range m {
		p := ext.Priority
		if _, exist := pMap[p]; exist {
			err := fmt.Errorf("duplicate priority %d on %s", p, n)
			return nil, err
		}
		pMap[p] = ext
	}

	var priorities []int
	for p := range pMap {
		priorities = append(priorities, p)
	}

	sort.Ints(priorities)

	var result []Extension
	for _, p := range priorities {
		result = append(result, pMap[p])
	}

	return result, nil
}

func pipeExtensions(reader io.ReadCloser, oid string, fileName string, extensions []Extension) (hash string, tmp *os.File, exts []PointerExtension, err error) {
	var extcmds []*extCommand
	for _, e := range extensions {
		ext := NewPointerExtension(e.Name, e.Priority, "")
		pieces := strings.Split(e.Clean, " ")
		name := strings.Trim(pieces[0], " ")
		var args []string
		for _, value := range pieces[1:] {
			arg := strings.Replace(value, "%f", fileName, -1)
			args = append(args, arg)
		}
		cmd := exec.Command(name, args...)
		ec := &extCommand{ext: ext, cmd: cmd}
		extcmds = append(extcmds, ec)
	}

	var input io.ReadCloser
	var output io.WriteCloser
	input = reader
	if tmp, err = TempFile(""); err != nil {
		return
	}
	defer tmp.Close()
	output = tmp

	last := len(extcmds) - 1
	for i, ec := range extcmds {
		ec.hasher = sha256.New()

		if i == last {
			ec.cmd.Stdout = io.MultiWriter(ec.hasher, output)
			ec.out = output
			continue
		}

		nextec := extcmds[i+1]
		var nextStdin io.WriteCloser
		var stdout io.ReadCloser
		if nextStdin, err = nextec.cmd.StdinPipe(); err != nil {
			return
		}
		if stdout, err = ec.cmd.StdoutPipe(); err != nil {
			return
		}

		ec.cmd.Stdin = input
		ec.cmd.Stdout = io.MultiWriter(ec.hasher, nextStdin)
		ec.out = nextStdin

		input = stdout

		var errBuff bytes.Buffer
		ec.err = &errBuff
		ec.cmd.Stderr = ec.err
	}

	for _, ec := range extcmds {
		if err = ec.cmd.Start(); err != nil {
			return
		}
	}

	for _, ec := range extcmds {
		if err = ec.cmd.Wait(); err != nil {
			err = fmt.Errorf("Extension '%s' failed with: %s", ec.ext.Name, ec.err.String())
			return
		}
		if err = ec.out.Close(); err != nil {
			return
		}
	}

	for _, ec := range extcmds {
		ec.ext.Oid = oid
		exts = append(exts, *ec.ext)
		if ec.cmd != nil {
			oid = hex.EncodeToString(ec.hasher.Sum(nil))
		}
	}

	hash = oid
	return
}
