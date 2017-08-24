package lfs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
)

// An entry from ls-tree or rev-list including a blob sha and tree path
type TreeBlob struct {
	Sha1     string
	Filename string
}

func runScanTree(cb GitScannerFoundPointer, ref string, filter *filepathfilter.Filter) error {
	// We don't use the nameMap approach here since that's imprecise when >1 file
	// can be using the same content
	treeShas, err := lsTreeBlobs(ref, filter)
	if err != nil {
		return err
	}

	pcw, err := catFileBatchTree(treeShas)
	if err != nil {
		return err
	}

	for p := range pcw.Results {
		cb(p, nil)
	}

	if err := pcw.Wait(); err != nil {
		cb(nil, err)
	}
	return nil
}

// catFileBatchTree uses git cat-file --batch to get the object contents
// of a git object, given its sha1. The contents will be decoded into
// a Git LFS pointer. treeblobs is a channel over which blob entries
// will be sent. It returns a channel from which point.Pointers can be read.
func catFileBatchTree(treeblobs *TreeBlobChannelWrapper) (*PointerChannelWrapper, error) {
	scanner, err := NewPointerScanner()
	if err != nil {
		scanner.Close()

		return nil, err
	}

	pointers := make(chan *WrappedPointer, chanBufSize)
	errchan := make(chan error, 10) // Multiple errors possible

	go func() {
		for t := range treeblobs.Results {
			hasNext := scanner.Scan(t.Sha1)
			if p := scanner.Pointer(); p != nil {
				p.Name = t.Filename
				pointers <- p
			}

			if err := scanner.Err(); err != nil {
				errchan <- err
			}

			if !hasNext {
				break
			}
		}

		// Deal with nested error from incoming treeblobs
		err := treeblobs.Wait()
		if err != nil {
			errchan <- err
		}

		if err = scanner.Close(); err != nil {
			errchan <- err
		}

		close(pointers)
		close(errchan)
	}()

	return NewPointerChannelWrapper(pointers, errchan), nil
}

// Use ls-tree at ref to find a list of candidate tree blobs which might be lfs files
// The returned channel will be sent these blobs which should be sent to catFileBatchTree
// for final check & conversion to Pointer
func lsTreeBlobs(ref string, filter *filepathfilter.Filter) (*TreeBlobChannelWrapper, error) {
	cmd, err := git.LsTree(ref)
	if err != nil {
		return nil, err
	}

	cmd.Stdin.Close()

	blobs := make(chan TreeBlob, chanBufSize)
	errchan := make(chan error, 1)

	go func() {
		scanner := newLsTreeScanner(cmd.Stdout)
		for scanner.Scan() {
			if t := scanner.TreeBlob(); t != nil && filter.Allows(t.Filename) {
				blobs <- *t
			}
		}

		stderr, _ := ioutil.ReadAll(cmd.Stderr)
		err := cmd.Wait()
		if err != nil {
			errchan <- fmt.Errorf("Error in git ls-tree: %v %v", err, string(stderr))
		}
		close(blobs)
		close(errchan)
	}()

	return NewTreeBlobChannelWrapper(blobs, errchan), nil
}

type lsTreeScanner struct {
	s    *bufio.Scanner
	tree *TreeBlob
}

func newLsTreeScanner(r io.Reader) *lsTreeScanner {
	s := bufio.NewScanner(r)
	s.Split(scanNullLines)
	return &lsTreeScanner{s: s}
}

func (s *lsTreeScanner) TreeBlob() *TreeBlob {
	return s.tree
}

func (s *lsTreeScanner) Err() error {
	return nil
}

func (s *lsTreeScanner) Scan() bool {
	t, hasNext := s.next()
	s.tree = t
	return hasNext
}

func (s *lsTreeScanner) next() (*TreeBlob, bool) {
	hasNext := s.s.Scan()
	line := s.s.Text()
	parts := strings.SplitN(line, "\t", 2)
	if len(parts) < 2 {
		return nil, hasNext
	}

	attrs := strings.SplitN(parts[0], " ", 4)
	if len(attrs) < 4 {
		return nil, hasNext
	}

	if attrs[1] != "blob" {
		return nil, hasNext
	}

	sz, err := strconv.ParseInt(strings.TrimSpace(attrs[3]), 10, 64)
	if err != nil {
		return nil, hasNext
	}

	if sz < blobSizeCutoff {
		sha1 := attrs[2]
		filename := parts[1]
		return &TreeBlob{Sha1: sha1, Filename: filename}, hasNext
	}
	return nil, hasNext
}

func scanNullLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\000'); i >= 0 {
		// We have a full null-terminated line.
		return i + 1, data[0:i], nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}
