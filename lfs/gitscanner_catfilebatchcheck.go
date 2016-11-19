package lfs

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
)

// runCatFileBatchCheck uses 'git cat-file --batch-check' to get the type and
// size of a git object. Any object that isn't of type blob and under the
// blobSizeCutoff will be ignored. revs is a channel over which strings
// containing git sha1s will be sent. It returns a channel from which sha1
// strings can be read.
func runCatFileBatchCheck(smallRevCh chan string, revs *StringChannelWrapper, errCh chan error) error {
	cmd, err := startCommand("git", "cat-file", "--batch-check")
	if err != nil {
		return err
	}

	go catFileBatchCheckOutput(smallRevCh, cmd, errCh)
	go catFileBatchCheckInput(cmd, revs, errCh)
	return nil
}

func catFileBatchCheckOutput(smallRevCh chan string, cmd *wrappedCmd, errCh chan error) {
	scanner := &catFileBatchCheckScanner{s: bufio.NewScanner(cmd.Stdout)}
	for scanner.Scan() {
		smallRevCh <- scanner.BlobOID()
	}

	if err := scanner.Err(); err != nil {
		errCh <- err
	}

	stderr, _ := ioutil.ReadAll(cmd.Stderr)
	err := cmd.Wait()
	if err != nil {
		errCh <- fmt.Errorf("Error in git cat-file --batch-check: %v %v", err, string(stderr))
	}
	close(smallRevCh)
	close(errCh)
}

func catFileBatchCheckInput(cmd *wrappedCmd, revs *StringChannelWrapper, errCh chan error) {
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

type catFileBatchCheckScanner struct {
	s       *bufio.Scanner
	blobOID string
	err     error
}

func (s *catFileBatchCheckScanner) BlobOID() string {
	return s.blobOID
}

func (s *catFileBatchCheckScanner) Err() error {
	return s.err
}

func (s *catFileBatchCheckScanner) Scan() bool {
	s.blobOID, s.err = "", nil
	b, err := scanBlobOID(s.s)
	if err != nil {
		// EOF halts scanning, but isn't a reportable error
		if err != io.EOF {
			s.err = err
		}
		return false
	}

	s.blobOID = b
	return true
}

func scanBlobOID(s *bufio.Scanner) (string, error) {
	objType := "blob"
	for s.Scan() {
		line := s.Text()
		lineLen := len(line)

		// Format is:
		// <sha1> <type> <size>
		// type is at a fixed spot, if we see that it's "blob", we can avoid
		// splitting the line just to get the size.
		if lineLen < 46 {
			continue
		}

		if line[41:45] != objType {
			continue
		}

		size, err := strconv.Atoi(line[46:lineLen])
		if err != nil {
			continue
		}

		if size < blobSizeCutoff {
			return line[0:40], nil
		}
	}

	return "", io.EOF
}
