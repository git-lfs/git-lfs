package lfs

import (
	"fmt"
	"io/ioutil"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
)

func runScanTree(cb GitScannerFoundPointer, ref string, filter *filepathfilter.Filter, gitEnv, osEnv config.Environment) error {
	// We don't use the nameMap approach here since that's imprecise when >1 file
	// can be using the same content
	treeShas, err := lsTreeBlobs(ref, filter)
	if err != nil {
		return err
	}

	pcw, err := catFileBatchTree(treeShas, gitEnv, osEnv)
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
func catFileBatchTree(treeblobs *TreeBlobChannelWrapper, gitEnv, osEnv config.Environment) (*PointerChannelWrapper, error) {
	scanner, err := NewPointerScanner(gitEnv, osEnv)
	if err != nil {
		return nil, err
	}

	pointers := make(chan *WrappedPointer, chanBufSize)
	errchan := make(chan error, 10) // Multiple errors possible

	go func() {
		hasNext := true
		for t := range treeblobs.Results {
			hasNext = scanner.Scan(t.Oid)

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

		// If the scanner quit early, we may still have treeblobs to
		// read, so waiting for it to close will cause a deadlock.
		if hasNext {
			// Deal with nested error from incoming treeblobs
			err := treeblobs.Wait()
			if err != nil {
				errchan <- err
			}
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

	blobs := make(chan git.TreeBlob, chanBufSize)
	errchan := make(chan error, 1)

	go func() {
		scanner := git.NewLsTreeScanner(cmd.Stdout)
		for scanner.Scan() {
			if t := scanner.TreeBlob(); t != nil && t.Size < blobSizeCutoff && filter.Allows(t.Filename) {
				blobs <- *t
			}
		}

		stderr, _ := ioutil.ReadAll(cmd.Stderr)
		err := cmd.Wait()
		if err != nil {
			errchan <- fmt.Errorf("error in git ls-tree: %v %v", err, string(stderr))
		}
		close(blobs)
		close(errchan)
	}()

	return NewTreeBlobChannelWrapper(blobs, errchan), nil
}
