package lfs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/git-lfs/git-lfs/errors"
)

// runCatFileBatch uses 'git cat-file --batch' to get the object contents of a
// git object, given its sha1. The contents will be decoded into a Git LFS
// pointer. Git Blob SHA1s are read from the sha1Ch channel and fed to STDIN.
// Results are parsed from STDOUT, and any elegible LFS pointers are sent to
// pointerCh. Any errors are sent to errCh. An error is returned if the 'git
// cat-file' command fails to start.
func runCatFileBatch(pointerCh chan *WrappedPointer, sha1Ch <-chan string, errCh chan error) error {
	cmd, err := startCommand("git", "cat-file", "--batch")
	if err != nil {
		return err
	}

	go catFileBatchOutput(pointerCh, cmd, errCh)
	go catFileBatchInput(cmd, sha1Ch, errCh)
	return nil
}

func catFileBatchOutput(pointerCh chan *WrappedPointer, cmd *wrappedCmd, errCh chan error) {
	scanner := &catFileBatchScanner{r: cmd.Stdout}
	for scanner.Scan() {
		pointerCh <- scanner.Pointer()
	}

	if err := scanner.Err(); err != nil {
		errCh <- err
	}

	stderr, _ := ioutil.ReadAll(cmd.Stderr)
	err := cmd.Wait()
	if err != nil {
		errCh <- fmt.Errorf("Error in git cat-file --batch: %v %v", err, string(stderr))
	}

	close(pointerCh)
	close(errCh)
}

func catFileBatchInput(cmd *wrappedCmd, sha1Ch <-chan string, errCh chan error) {
	for r := range sha1Ch {
		cmd.Stdin.Write([]byte(r + "\n"))
	}
	cmd.Stdin.Close()
}

type catFileBatchScanner struct {
	r       *bufio.Reader
	pointer *WrappedPointer
	err     error
}

func (s *catFileBatchScanner) Pointer() *WrappedPointer {
	return s.pointer
}

func (s *catFileBatchScanner) Err() error {
	return s.err
}

func (s *catFileBatchScanner) Scan() bool {
	s.pointer, s.err = nil, nil
	p, err := scanPointer(s.r)
	if err != nil && err != io.EOF {
		s.err = err
		return false
	}

	s.pointer = p
	return true
}

func scanPointer(r *bufio.Reader) (*WrappedPointer, error) {
	var pointer *WrappedPointer

	for pointer == nil {
		l, err := r.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}

		// Line is formatted:
		// <sha1> <type> <size>
		fields := bytes.Fields(l)
		if len(fields) < 3 {
			return nil, errors.Wrap(fmt.Errorf("Invalid: %s", string(l)), "git cat-file --batch:")
		}

		size, _ := strconv.Atoi(string(fields[2]))
		p, err := DecodePointer(io.LimitReader(r, int64(size)))
		if err == nil {
			pointer = &WrappedPointer{
				Sha1:    string(fields[0]),
				Size:    p.Size,
				Pointer: p,
			}
		}

		_, err = r.ReadBytes('\n') // Extra \n inserted by cat-file
		if err != nil && err != io.EOF {
			return nil, err
		}
	}

	return pointer, nil
}
