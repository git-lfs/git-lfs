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
// Results are parsed from STDOUT, and any eligible LFS pointers are sent to
// pointerCh. If a Git Blob is not an LFS pointer, check the lockableSet to see
// if that blob is for a locked file. Any errors are sent to errCh. An error is
// returned if the 'git cat-file' command fails to start.
func runCatFileBatch(pointerCh chan *WrappedPointer, lockableCh chan string, lockableSet *lockableNameSet, revs *StringChannelWrapper, errCh chan error) error {
	cmd, err := startCommand("git", "cat-file", "--batch")
	if err != nil {
		return err
	}

	go func() {
		scanner := &catFileBatchScanner{r: cmd.Stdout}
		for r := range revs.Results {
			cmd.Stdin.Write([]byte(r + "\n"))
			canScan := scanner.Scan()

			if err := scanner.Err(); err != nil {
				errCh <- err
			} else if p := scanner.Pointer(); p != nil {
				pointerCh <- p
			} else if b := scanner.BlobSHA(); len(b) == 40 {
				if name, ok := lockableSet.Check(b); ok {
					lockableCh <- name
				}
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
		close(lockableCh)
	}()

	return nil
}

type catFileBatchScanner struct {
	r       *bufio.Reader
	blobSha string
	pointer *WrappedPointer
	err     error
}

func (s *catFileBatchScanner) BlobSHA() string {
	return s.blobSha
}

func (s *catFileBatchScanner) Pointer() *WrappedPointer {
	return s.pointer
}

func (s *catFileBatchScanner) Err() error {
	return s.err
}

func (s *catFileBatchScanner) Scan() bool {
	s.pointer, s.err = nil, nil
	b, p, err := s.next()
	s.blobSha = b
	s.pointer = p

	if err != nil {
		if err != io.EOF {
			s.err = err
		}
		return false
	}

	return true
}

func (s *catFileBatchScanner) next() (string, *WrappedPointer, error) {
	l, err := s.r.ReadBytes('\n')
	if err != nil {
		return "", nil, err
	}

	// Line is formatted:
	// <sha1> <type> <size>
	fields := bytes.Fields(l)
	if len(fields) < 3 {
		return "", nil, errors.Wrap(fmt.Errorf("Invalid: %q", string(l)), "git cat-file --batch")
	}

	blobSha := string(fields[0])
	size, _ := strconv.Atoi(string(fields[2]))
	buf := make([]byte, size)
	read, err := io.ReadFull(s.r, buf)
	if err != nil {
		return blobSha, nil, err
	}

	if size != read {
		return blobSha, nil, fmt.Errorf("expected %d bytes, read %d bytes", size, read)
	}

	p, err := DecodePointer(bytes.NewBuffer(buf[:read]))
	var pointer *WrappedPointer
	if err == nil {
		pointer = &WrappedPointer{
			Sha1:    blobSha,
			Pointer: p,
		}
	}

	_, err = s.r.ReadBytes('\n') // Extra \n inserted by cat-file
	return blobSha, pointer, err
}
