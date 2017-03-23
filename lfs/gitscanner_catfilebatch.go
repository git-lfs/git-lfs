package lfs

import (
	"bufio"
	"bytes"
	"crypto/sha256"
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
	scanner, err := NewCatFileBatchScanner()
	if err != nil {
		return err
	}

	go func() {
		for r := range revs.Results {
			canScan := scanner.Scan([]byte(r))

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

		if err := scanner.Close(); err != nil {
			errCh <- err
		}

		close(pointerCh)
		close(errCh)
		close(lockableCh)
	}()

	return nil
}

type CatFileBatchScanner struct {
	r       *bufio.Reader
	w       io.Writer
	closeFn func() error

	blobSha     string
	contentsSha string
	pointer     *WrappedPointer
	err         error
}

func NewCatFileBatchScanner() (*CatFileBatchScanner, error) {
	cmd, err := startCommand("git", "cat-file", "--batch")
	if err != nil {
		return nil, err
	}

	closeFn := func() error {
		if err := cmd.Stdin.Close(); err != nil {
			return err
		}

		stderr, _ := ioutil.ReadAll(cmd.Stderr)
		if err := cmd.Wait(); err != nil {
			return errors.Errorf("Error in git cat-file --batch: %v %v", err, string(stderr))
		}

		return nil
	}

	return &CatFileBatchScanner{
		r:       cmd.Stdout,
		w:       cmd.Stdin,
		closeFn: closeFn,
	}, nil
}

func (s *CatFileBatchScanner) BlobSHA() string {
	return s.blobSha
}

func (s *CatFileBatchScanner) ContentsSha() string {
	return s.contentsSha
}

func (s *CatFileBatchScanner) Pointer() *WrappedPointer {
	return s.pointer
}

func (s *CatFileBatchScanner) Err() error {
	return s.err
}

func (s *CatFileBatchScanner) Scan(sha []byte) bool {
	s.pointer, s.err = nil, nil
	s.blobSha, s.contentsSha = "", ""

	if s.w != nil && sha != nil {
		if _, err := fmt.Fprintf(s.w, "%s\n", sha); err != nil {
			s.err = err
			return false
		}
	}

	b, c, p, err := s.next()
	s.blobSha = b
	s.contentsSha = c
	s.pointer = p

	if err != nil {
		if err != io.EOF {
			s.err = err
		}
		return false
	}

	return true
}

func (s *CatFileBatchScanner) Close() error {
	if s.closeFn == nil {
		return nil
	}
	return s.closeFn()
}

func (s *CatFileBatchScanner) next() (string, string, *WrappedPointer, error) {
	l, err := s.r.ReadBytes('\n')
	if err != nil {
		return "", "", nil, err
	}

	// Line is formatted:
	// <sha1> <type> <size>
	fields := bytes.Fields(l)
	if len(fields) < 3 {
		return "", "", nil, errors.Wrap(fmt.Errorf("Invalid: %q", string(l)), "git cat-file --batch")
	}

	blobSha := string(fields[0])
	size, _ := strconv.Atoi(string(fields[2]))
	sha := sha256.New()

	var buf *bytes.Buffer
	var to io.Writer = sha
	if size <= blobSizeCutoff {
		buf = bytes.NewBuffer(make([]byte, 0, size))
		to = io.MultiWriter(to, buf)
	}

	read, err := io.CopyN(to, s.r, int64(size))
	if err != nil {
		return blobSha, "", nil, err
	}

	if int64(size) != read {
		return blobSha, "", nil, fmt.Errorf("expected %d bytes, read %d bytes", size, read)
	}

	var pointer *WrappedPointer
	var contentsSha string

	if size <= blobSizeCutoff {
		if p, err := DecodePointer(bytes.NewReader(buf.Bytes())); err != nil {
			contentsSha = fmt.Sprintf("%x", sha.Sum(nil))
		} else {
			pointer = &WrappedPointer{
				Sha1:    blobSha,
				Pointer: p,
			}
			contentsSha = p.Oid
		}
	} else {
		contentsSha = fmt.Sprintf("%x", sha.Sum(nil))
	}

	_, err = s.r.ReadBytes('\n') // Extra \n inserted by cat-file

	return blobSha, contentsSha, pointer, err
}
