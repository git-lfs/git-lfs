package lfs

import (
	"bufio"
	"fmt"
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

	go func() {
		scanner := &catFileBatchCheckScanner{s: bufio.NewScanner(cmd.Stdout), limit: blobSizeCutoff}
		for r := range revs.Results {
			cmd.Stdin.Write([]byte(r + "\n"))
			hasNext := scanner.Scan()
			if b := scanner.BlobOID(); len(b) > 0 {
				smallRevCh <- b
			}

			if err := scanner.Err(); err != nil {
				errCh <- err
			}

			if !hasNext {
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
			errCh <- fmt.Errorf("Error in git cat-file --batch-check: %v %v", err, string(stderr))
		}
		close(smallRevCh)
		close(errCh)
	}()

	return nil
}

type catFileBatchCheckScanner struct {
	s       *bufio.Scanner
	limit   int
	blobOID string
}

func (s *catFileBatchCheckScanner) BlobOID() string {
	return s.blobOID
}

func (s *catFileBatchCheckScanner) Err() error {
	return s.s.Err()
}

func (s *catFileBatchCheckScanner) Scan() bool {
	s.blobOID = ""
	b, hasNext := s.next()
	s.blobOID = b
	return hasNext
}

func (s *catFileBatchCheckScanner) next() (string, bool) {
	hasNext := s.s.Scan()
	line := s.s.Text()
	lineLen := len(line)

	// Format is:
	// <sha1> <type> <size>
	// type is at a fixed spot, if we see that it's "blob", we can avoid
	// splitting the line just to get the size.
	if lineLen < 46 {
		return "", hasNext
	}

	if line[41:45] != "blob" {
		return "", hasNext
	}

	size, err := strconv.Atoi(line[46:lineLen])
	if err != nil {
		return "", hasNext
	}

	if size >= s.limit {
		return "", hasNext
	}

	return line[0:40], hasNext
}
