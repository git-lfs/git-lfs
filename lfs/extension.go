package lfs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/git-lfs/git-lfs/config"
)

type pipeRequest struct {
	action     string
	reader     io.Reader
	fileName   string
	extensions []config.Extension
}

type pipeResponse struct {
	file    *os.File
	results []*pipeExtResult
}

type pipeExtResult struct {
	name   string
	oidIn  string
	oidOut string
}

type extCommand struct {
	cmd    *exec.Cmd
	out    io.WriteCloser
	err    *bytes.Buffer
	hasher hash.Hash
	result *pipeExtResult
}

func pipeExtensions(cfg *config.Configuration, request *pipeRequest) (response pipeResponse, err error) {
	var extcmds []*extCommand
	defer func() {
		// In the case of an early return before the end of this
		// function (in response to an error, etc), kill all running
		// processes. Errors are ignored since the function has already
		// returned.
		//
		// In the happy path, the commands will have already been
		// `Wait()`-ed upon and e.cmd.Process.Kill() will return an
		// error, but we can ignore it.
		for _, e := range extcmds {
			if e.cmd.Process != nil {
				e.cmd.Process.Kill()
			}
		}
	}()

	for _, e := range request.extensions {
		var pieces []string
		switch request.action {
		case "clean":
			pieces = strings.Split(e.Clean, " ")
		case "smudge":
			pieces = strings.Split(e.Smudge, " ")
		default:
			err = fmt.Errorf("Invalid action: " + request.action)
			return
		}
		name := strings.Trim(pieces[0], " ")
		var args []string
		for _, value := range pieces[1:] {
			arg := strings.Replace(value, "%f", request.fileName, -1)
			args = append(args, arg)
		}
		cmd := exec.Command(name, args...)
		ec := &extCommand{cmd: cmd, result: &pipeExtResult{name: e.Name}}
		extcmds = append(extcmds, ec)
	}

	hasher := sha256.New()
	pipeReader, pipeWriter := io.Pipe()
	multiWriter := io.MultiWriter(hasher, pipeWriter)

	var input io.Reader
	var output io.WriteCloser
	input = pipeReader
	extcmds[0].cmd.Stdin = input
	if response.file, err = ioutil.TempFile(cfg.TempDir(), ""); err != nil {
		return
	}
	defer response.file.Close()
	output = response.file

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

	if _, err = io.Copy(multiWriter, request.reader); err != nil {
		return
	}
	if err = pipeWriter.Close(); err != nil {
		return
	}

	for _, ec := range extcmds {
		if err = ec.cmd.Wait(); err != nil {
			if ec.err != nil {
				errStr := ec.err.String()
				err = fmt.Errorf("Extension '%s' failed with: %s", ec.result.name, errStr)
			}
			return
		}
		if err = ec.out.Close(); err != nil {
			return
		}
	}

	oid := hex.EncodeToString(hasher.Sum(nil))
	for _, ec := range extcmds {
		ec.result.oidIn = oid
		oid = hex.EncodeToString(ec.hasher.Sum(nil))
		ec.result.oidOut = oid
		response.results = append(response.results, ec.result)
	}
	return
}
