package lfs

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/git-lfs/git-lfs/git"
)

// runCatFileBatchCheck uses 'git cat-file --batch-check' to get the type and
// size of a git object. Any object that isn't of type blob and under the
// blobSizeCutoff will be ignored, unless it's a locked file. revs is a channel
// over which strings containing git sha1s will be sent. It returns a channel
// from which sha1 strings can be read.
func runCatFileBatchCheck(smallRevCh chan string, lockableCh chan string, lockableSet *lockableNameSet, revs *StringChannelWrapper, errCh chan error) error {
	cmd, err := git.CatFile()
	if err != nil {
		return err
	}

	go func() {
		scanner := &catFileBatchCheckScanner{s: bufio.NewScanner(cmd.Stdout), limit: blobSizeCutoff}
		for r := range revs.Results {
			cmd.Stdin.Write([]byte(r + "\n"))
			hasNext := scanner.Scan()
			if err := scanner.Err(); err != nil {
				errCh <- err
			} else if b := scanner.LFSBlobOID(); len(b) > 0 {
				smallRevCh <- b
			} else if b := scanner.GitBlobOID(); len(b) > 0 {
				if name, ok := lockableSet.Check(b); ok {
					lockableCh <- name
				}
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
			errCh <- fmt.Errorf("error in git cat-file --batch-check: %v %v", err, string(stderr))
		}
		close(smallRevCh)
		close(errCh)
	}()

	return nil
}

type catFileBatchCheckScanner struct {
	s          *bufio.Scanner
	limit      int
	lfsBlobOID string
	gitBlobOID string
}

func (s *catFileBatchCheckScanner) LFSBlobOID() string {
	return s.lfsBlobOID
}

func (s *catFileBatchCheckScanner) GitBlobOID() string {
	return s.gitBlobOID
}

func (s *catFileBatchCheckScanner) Err() error {
	return s.s.Err()
}

func (s *catFileBatchCheckScanner) Scan() bool {
	lfsBlobSha, gitBlobSha, hasNext := s.next()
	s.lfsBlobOID = lfsBlobSha
	s.gitBlobOID = gitBlobSha
	return hasNext
}

func (s *catFileBatchCheckScanner) next() (string, string, bool) {
	hasNext := s.s.Scan()
	line := s.s.Text()
	lineLen := len(line)

	oidLen := strings.IndexByte(line, ' ')

	// Format is:
	// <hash> <type> <size>
	// type is at a fixed spot, if we see that it's "blob", we can avoid
	// splitting the line just to get the size.
	if oidLen == -1 || lineLen < oidLen+6 {
		return "", "", hasNext
	}

	if line[oidLen+1:oidLen+5] != "blob" {
		return "", "", hasNext
	}

	size, err := strconv.Atoi(line[oidLen+6 : lineLen])
	if err != nil {
		return "", "", hasNext
	}

	blobSha := line[0:oidLen]
	if size >= s.limit {
		return "", blobSha, hasNext
	}

	return blobSha, "", hasNext
}
