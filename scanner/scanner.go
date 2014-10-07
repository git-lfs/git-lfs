package scanner

import (
	"bufio"
	"bytes"
	"github.com/github/git-media/pointer"
	// "github.com/rubyist/tracerx"
	"io"
	"os/exec"
	"strconv"
)

var (
	blobSizeCutoff = 125
	stdoutBufSize  = 16384
)

func Scan(ref string) ([]*pointer.Pointer, error) {
	revs, _ := revListStream(ref, ref == "")
	smallShas, _ := catFileBatchCheck(revs)
	pointerc, _ := catFileBatch(smallShas)

	pointers := make([]*pointer.Pointer, 0)
	for p := range pointerc {
		pointers = append(pointers, p)
	}

	return pointers, nil
}

func revListStream(ref string, all bool) (chan string, error) {
	refArgs := []string{"rev-list", "--objects"}
	if all {
		refArgs = append(refArgs, "--all")
	} else {
		refArgs = append(refArgs, ref)
	}

	cmd, err := startCommand("git", refArgs...)
	if err != nil {
		return nil, err
	}

	cmd.Stdin.Close()

	revs := make(chan string)

	go func() {
		scanner := bufio.NewScanner(cmd.Stdout)
		for scanner.Scan() {
			revs <- scanner.Text()[0:40]
		}
		close(revs)
	}()

	return revs, nil
}

func catFileBatchCheck(revs chan string) (chan string, error) {
	cmd, err := startCommand("git", "cat-file", "--batch-check")
	if err != nil {
		return nil, err
	}

	smallRevs := make(chan string)

	go func() {
		scanner := bufio.NewScanner(cmd.Stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if line[41:45] == "blob" {
				size, err := strconv.Atoi(line[46:len(line)])
				if err != nil {
					continue
				}
				if size < blobSizeCutoff {
					smallRevs <- line[0:40]
				}
			}
		}
		close(smallRevs)
	}()

	go func() {
		for r := range revs {
			cmd.Stdin.Write([]byte(r + "\n"))
		}
		cmd.Stdin.Close()
	}()

	return smallRevs, nil
}

func catFileBatch(revs chan string) (chan *pointer.Pointer, error) {
	cmd, err := startCommand("git", "cat-file", "--batch")
	if err != nil {
		return nil, err
	}

	pointers := make(chan *pointer.Pointer)

	go func() {
		for {
			l, err := cmd.Stdout.ReadBytes('\n')
			if err != nil { // Probably check for EOF
				break
			}

			// Line is formatted:
			// <sha1> <type> <size>
			fields := bytes.Fields(l)
			s, _ := strconv.Atoi(string(fields[2]))

			nbuf := make([]byte, s)
			_, err = io.ReadFull(cmd.Stdout, nbuf)
			if err != nil {
				break // Legit errors
			}

			p, err := pointer.Decode(bytes.NewBuffer(nbuf))
			if err == nil {
				pointers <- p
			}

			_, err = cmd.Stdout.ReadBytes('\n') // Extra \n inserted by cat-file
			if err != nil {                     // Probably check for EOF
				break
			}
		}
		close(pointers)
	}()

	// writes shas to cat-file stdin
	go func() {
		for r := range revs {
			cmd.Stdin.Write([]byte(r + "\n"))
		}
		cmd.Stdin.Close()
	}()

	return pointers, nil
}

type wrappedCmd struct {
	Stdin  io.WriteCloser
	Stdout *bufio.Reader
	*exec.Cmd
}

func startCommand(command string, args ...string) (*wrappedCmd, error) {
	cmd := exec.Command(command, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &wrappedCmd{stdin, bufio.NewReaderSize(stdout, stdoutBufSize), cmd}, nil
}
