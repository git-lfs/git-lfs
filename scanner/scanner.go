package scanner

import (
	"bufio"
	"bytes"
	"github.com/github/git-media/pointer"
	"github.com/rubyist/tracerx"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	blobSizeCutoff = 130
	stdoutBufSize  = 16384
	chanBufSize    = 100
)

type wrappedPointer struct {
	Sha1 string
	Name string
	*pointer.Pointer
}

// Scan takes a ref and returns a slice of pointer.Pointer objects
// for all git media pointers it finds for that ref.
func Scan(ref string) ([]*wrappedPointer, error) {
	nameMap := make(map[string]string, 0)
	start := time.Now()

	revs, err := revListShas(ref, ref == "", nameMap)
	if err != nil {
		return nil, err
	}

	smallShas, err := catFileBatchCheck(revs)
	if err != nil {
		return nil, err
	}

	pointerc, err := catFileBatch(smallShas)
	if err != nil {
		return nil, err
	}

	pointers := make([]*wrappedPointer, 0)
	for p := range pointerc {
		if name, ok := nameMap[p.Sha1]; ok {
			p.Name = name
		}
		pointers = append(pointers, p)
	}

	tracerx.PerformanceSince("scan", start)

	return pointers, nil
}

// revListShas uses git rev-list to return the list of object sha1s
// for the given ref. If all is true, ref is ignored. It returns a
// channel from which sha1 strings can be read.
func revListShas(ref string, all bool, nameMap map[string]string) (chan string, error) {
	refArgs := []string{"rev-list", "--objects"}
	if all {
		refArgs = append(refArgs, "--all")
	} else {
		refArgs = append(refArgs, "--no-walk")
		refArgs = append(refArgs, ref)
	}

	cmd, err := startCommand("git", refArgs...)
	if err != nil {
		return nil, err
	}

	cmd.Stdin.Close()

	revs := make(chan string, chanBufSize)

	go func() {
		scanner := bufio.NewScanner(cmd.Stdout)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			sha1 := line[0:40]
			if len(line) > 40 {
				nameMap[sha1] = line[41:len(line)]
			}
			revs <- sha1
		}
		close(revs)
	}()

	return revs, nil
}

// catFileBatchCheck uses git cat-file --batch-check to get the type
// and size of a git object. Any object that isn't of type blob and
// under the blobSizeCutoff will be ignored. revs is a channel over
// which strings containing git sha1s will be sent. It returns a channel
// from which sha1 strings can be read.
func catFileBatchCheck(revs chan string) (chan string, error) {
	cmd, err := startCommand("git", "cat-file", "--batch-check")
	if err != nil {
		return nil, err
	}

	smallRevs := make(chan string, chanBufSize)

	go func() {
		scanner := bufio.NewScanner(cmd.Stdout)
		for scanner.Scan() {
			line := scanner.Text()
			// Format is:
			// <sha1> <type> <size>
			// type is at a fixed spot, if we see that it's "blob", we can avoid
			// splitting the line just to get the size.
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

// catFileBatch uses git cat-file --batch to get the object contents
// of a git object, given its sha1. The contents will be decoded into
// a git media pointer. revs is a channel over which strings containing
// git sha1s will be sent. It returns a channel from which point.Pointers
// can be read.
func catFileBatch(revs chan string) (chan *wrappedPointer, error) {
	cmd, err := startCommand("git", "cat-file", "--batch")
	if err != nil {
		return nil, err
	}

	pointers := make(chan *wrappedPointer, chanBufSize)

	go func() {
		for {
			l, err := cmd.Stdout.ReadBytes('\n')
			if err != nil {
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
				pointers <- &wrappedPointer{string(fields[0]), "", p}
			}

			_, err = cmd.Stdout.ReadBytes('\n') // Extra \n inserted by cat-file
			if err != nil {
				break
			}
		}
		close(pointers)
	}()

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

// startCommand starts up a command and creates a stdin pipe and a buffered
// stdout pipe, wrapped in a wrappedCmd. The stdout buffer wille be of stdoutBufSize
// bytes.
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

	tracerx.Printf("run_command: %s %s", command, strings.Join(args, " "))
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &wrappedCmd{stdin, bufio.NewReaderSize(stdout, stdoutBufSize), cmd}, nil
}
