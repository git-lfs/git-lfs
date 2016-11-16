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

	go catFileBatchCheckOutput(smallRevCh, cmd, errCh)
	go catFileBatchCheckInput(cmd, revs, errCh)
	return nil
}

func catFileBatchCheckOutput(smallRevCh chan string, cmd *wrappedCmd, errCh chan error) {
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
			smallRevCh <- line[0:40]
		}
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
