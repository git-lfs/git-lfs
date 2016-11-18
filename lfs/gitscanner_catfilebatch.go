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
func runCatFileBatch(pointerCh chan *WrappedPointer, revs *StringChannelWrapper, errCh chan error) error {
	cmd, err := startCommand("git", "cat-file", "--batch")
	if err != nil {
		return err
	}

	go catFileBatchOutput(pointerCh, cmd, errCh)
	go catFileBatchInput(cmd, revs, errCh)
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

func catFileBatchInput(cmd *wrappedCmd, revs *StringChannelWrapper, errCh chan error) {
	for r := range revs.Results {
		cmd.Stdin.Write([]byte(r + "\n"))
	}
	err := revs.Wait()
	if err != nil {
		// We can share errchan with other goroutine since that won't close it
		// until we close the stdin below
		errCh <- err
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

func (s *catFileBatchScanner) Next() (*WrappedPointer, error) {
	return scanChunk(s.r)
}

func (s *catFileBatchScanner) Scan() bool {
	s.pointer, s.err = nil, nil
	p, err := scanPointer(s.r)
	if err != nil {
		// EOF halts scanning, but isn't a reportable error
		if err != io.EOF {
			s.err = err
		}
		return false
	}

	s.pointer = p
	return true
}

func scanChunk(r *bufio.Reader) (*WrappedPointer, error) {
	l, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	// Line is formatted:
	// <sha1> <type> <size>
	fields := bytes.Fields(l)
	if len(fields) < 3 {
		return nil, errors.Wrap(fmt.Errorf("Invalid: %q", string(l)), "git cat-file --batch")
	}

	size, _ := strconv.Atoi(string(fields[2]))
	buf := make([]byte, size)
	read, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	if size != read {
		return nil, fmt.Errorf("expected %d bytes, read %d bytes", size, read)
	}

	p, err := DecodePointer(bytes.NewBuffer(buf[0:read]))
	var pointer *WrappedPointer
	if err == nil {
		pointer = &WrappedPointer{
			Sha1:    string(fields[0]),
			Pointer: p,
		}
	}

	_, err = r.ReadBytes('\n') // Extra \n inserted by cat-file
	return pointer, err
}

func scanPointer(r *bufio.Reader) (*WrappedPointer, error) {
	var pointer *WrappedPointer

	for pointer == nil {
		p, err := scanChunk(r)
		if err != nil {
			return nil, err
		}
		pointer = p
	}

	return pointer, nil
}
