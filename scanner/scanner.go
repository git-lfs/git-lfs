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

type ScannedPointer struct {
	Name string
	*pointer.Pointer
}

func revListStream(ref string, all bool) (chan string, error) {
	refArgs := []string{"rev-list", "--objects"}
	if all {
		refArgs = append(refArgs, "--all")
	} else {
		refArgs = append(refArgs, ref)
	}

	cmd := exec.Command("git", refArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	tracerx.Printf("run_command: 'git' %s", strings.Join(refArgs, " "))
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	revs := make(chan string)

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			revs <- scanner.Text()[0:40]
		}
		close(revs)
	}()

	return revs, nil
}

func catFileBatchCheck(revs chan string) (chan string, error) {
	cmd := exec.Command("git", "cat-file", "--batch-check")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	smallRevs := make(chan string)

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if line[41:45] == "blob" {
				size, err := strconv.Atoi(line[46:len(line)])
				if err != nil {
					continue
				}
				if size < 200 {
					smallRevs <- line[0:40]
				}
			}
		}
		close(smallRevs)
	}()

	go func() {
		for r := range revs {
			stdin.Write([]byte(r + "\n"))
		}
		stdin.Close()
	}()

	return smallRevs, nil
}

func catFileBatch(revs chan string) (chan *pointer.Pointer, error) {
	cmd := exec.Command("git", "cat-file", "--batch")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	pointers := make(chan *pointer.Pointer)

	// reads from cat-file stdout, parses out pointers
	go func() {
		bstdout := bufio.NewReader(stdout)
		for {
			l, err := bstdout.ReadBytes('\n')
			if err != nil { // Probably check for EOF
				break
			}

			tracerx.Printf("l: .%s.", string(l))

			fields := bytes.Fields(l)
			s, _ := strconv.Atoi(string(fields[2]))

			nbuf := make([]byte, s)
			_, err = io.ReadFull(bstdout, nbuf)
			if err != nil {
				break // Legit errors
			}

			p, err := pointer.Decode(bytes.NewBuffer(nbuf))
			if err == nil {
				pointers <- p
			}

			_, err = bstdout.ReadBytes('\n') // Extra \n inserted by cat-file
			if err != nil {                  // Probably check for EOF
				break
			}
		}
		close(pointers)
	}()

	// writes shas to cat-file stdin
	go func() {
		for r := range revs {
			stdin.Write([]byte(r + "\n"))
		}
		stdin.Close()
	}()

	return pointers, nil
}
