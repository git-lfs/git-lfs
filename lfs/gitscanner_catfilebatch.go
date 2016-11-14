package lfs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/pkg/errors"
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
	for {
		l, err := cmd.Stdout.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				errCh <- errors.Wrap(err, "git cat-file --batch:")
			}
			break
		}

		// Line is formatted:
		// <sha1> <type> <size>
		fields := bytes.Fields(l)
		if len(fields) < 3 {
			panic("Bad fields??? " + string(l))
		}
		s, _ := strconv.Atoi(string(fields[2]))

		nbuf := make([]byte, s)
		_, err = io.ReadFull(cmd.Stdout, nbuf)
		if err != nil {
			break // Legit errors
		}

		p, err := DecodePointer(bytes.NewBuffer(nbuf))
		if err == nil {
			pointerCh <- &WrappedPointer{
				Sha1:    string(fields[0]),
				Size:    p.Size,
				Pointer: p,
			}
		}

		_, err = cmd.Stdout.ReadBytes('\n') // Extra \n inserted by cat-file
		if err != nil {
			if err != io.EOF {
				errCh <- errors.Wrap(err, "git cat-file --batch:")
			}
			break
		}
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
