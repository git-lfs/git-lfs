package scanner

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/github/git-lfs/tools"
)

const (
	// blobSizeCutoff is used to determine which files to scan for Git LFS
	// pointers.  Any file with a size below this cutoff will be scanned.
	blobSizeCutoff = 1024
)

// CatFileBatchCheck uses git cat-file --batch-check to get the type
// and size of a git object. Any object that isn't of type blob and
// under the blobSizeCutoff will be ignored. revs is a channel over
// which strings containing git sha1s will be sent. It returns a channel
// from which sha1 strings can be read.
func CatFileBatchCheck(revs <-chan string, errs <-chan error) (<-chan string, <-chan error, error) {
	cmd, err := startCommand("git", "cat-file", "--batch-check")
	if err != nil {
		return nil, nil, err
	}

	smallRevs := make(chan string, chanBufSize)
	errchan := make(chan error, 2) // up to 2 errors, one from each goroutine

	go func() {
		scanner := bufio.NewScanner(cmd.Stdout)
		for scanner.Scan() {
			line := scanner.Text()
			lineLen := len(line)

			// Format is:
			// <sha1> <type> <size>
			// type is at a fixed spot, if we see that it's "blob", we can avoid
			// splitting the line just to get the size.
			if lineLen < 46 {
				continue
			}

			if line[41:45] != "blob" {
				continue
			}

			size, err := strconv.Atoi(line[46:lineLen])
			if err != nil {
				continue
			}

			if size < blobSizeCutoff {
				smallRevs <- line[0:40]
			}
		}

		stderr, _ := ioutil.ReadAll(cmd.Stderr)
		err := cmd.Wait()
		if err != nil {
			errchan <- fmt.Errorf("Error in git cat-file --batch-check: %v %v", err, string(stderr))
		}
		close(smallRevs)
		close(errchan)
	}()

	go func() {
		for r := range revs {
			cmd.Stdin.Write([]byte(r + "\n"))
		}

		if err := tools.CollectErrsFromChan(errs); err != nil {
			// We can share errchan with other goroutine since that won't close it
			// until we close the stdin below
			errchan <- err
		}

		cmd.Stdin.Close()
	}()

	return smallRevs, errchan, nil
}
