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

	go func() {
		scanner := &catFileBatchScanner{r: cmd.Stdout}
		for r := range revs.Results {
			cmd.Stdin.Write([]byte(r + "\n"))
			canScan := scanner.Scan()
			if p := scanner.Pointer(); p != nil {
				pointerCh <- p
			}

			if err := scanner.Err(); err != nil {
				errCh <- err
			}

			if !canScan {
				break
			}
		}

		if err := revs.Wait(); err != nil {
			errCh <- err
		}

		cmd.Stdin.Close()

		stderr, _ := ioutil.ReadAll(cmd.Stderr)
		err := cmd.Wait()
		if err != nil {
			errCh <- fmt.Errorf("Error in git cat-file --batch: %v %v", err, string(stderr))
		}

		close(pointerCh)
		close(errCh)
	}()

	return nil
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
	p, err := s.next()
	s.pointer = p

	if err != nil {
		if err != io.EOF {
			s.err = err
		}
		return false
	}

	return true
}

func (s *catFileBatchScanner) next() (*WrappedPointer, error) {
	l, err := s.r.ReadBytes('\n')
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
	read, err := io.ReadFull(s.r, buf)
	if err != nil {
		return nil, err
	}

	if size != read {
		return nil, fmt.Errorf("expected %d bytes, read %d bytes", size, read)
	}

	p, err := DecodePointer(bytes.NewBuffer(buf[:read]))
	var pointer *WrappedPointer
	if err == nil {
		pointer = &WrappedPointer{
			Sha1:    string(fields[0]),
			Pointer: p,
		}
	}

	_, err = s.r.ReadBytes('\n') // Extra \n inserted by cat-file
	return pointer, err
}
