package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/subprocess"
	"github.com/git-lfs/git-lfs/tq"
)

// Handles the process of checking out a single file, and updating the git
// index.
func newSingleCheckout() *singleCheckout {
	// Get a converter from repo-relative to cwd-relative
	// Since writing data & calling git update-index must be relative to cwd
	pathConverter, err := lfs.NewRepoToCurrentPathConverter()
	if err != nil {
		Panic(err, "Could not convert file paths")
	}

	return &singleCheckout{
		gitIndexer:    &gitIndexer{},
		pathConverter: pathConverter,
		manifest:      getTransferManifest(),
	}
}

type singleCheckout struct {
	gitIndexer    *gitIndexer
	pathConverter lfs.PathConverter
	manifest      *tq.Manifest
}

func (c *singleCheckout) Run(p *lfs.WrappedPointer) {
	// Check the content - either missing or still this pointer (not exist is ok)
	filepointer, err := lfs.DecodePointerFromFile(p.Name)
	if err != nil && !os.IsNotExist(err) {
		if errors.IsNotAPointerError(err) {
			// File has non-pointer content, leave it alone
			return
		}

		LoggedError(err, "Checkout error: %s", err)
		return
	}

	if filepointer != nil && filepointer.Oid != p.Oid {
		// User has probably manually reset a file to another commit
		// while leaving it a pointer; don't mess with this
		return
	}

	cwdfilepath := c.pathConverter.Convert(p.Name)

	err = lfs.PointerSmudgeToFile(cwdfilepath, p.Pointer, false, c.manifest, nil)
	if err != nil {
		if errors.IsDownloadDeclinedError(err) {
			// acceptable error, data not local (fetch not run or include/exclude)
			LoggedError(err, "Skipped checkout for %q, content not local. Use fetch to download.", p.Name)
		} else {
			FullError(fmt.Errorf("Could not check out %q", p.Name))
		}
		return
	}

	// errors are only returned when the gitIndexer is starting a new cmd
	if err := c.gitIndexer.Add(cwdfilepath); err != nil {
		Panic(err, "Could not update the index")
	}
}

func (c *singleCheckout) Close() {
	if err := c.gitIndexer.Close(); err != nil {
		LoggedError(err, "Error updating the git index:\n%s", c.gitIndexer.Output())
	}
}

// Don't fire up the update-index command until we have at least one file to
// give it. Otherwise git interprets the lack of arguments to mean param-less update-index
// which can trigger entire working copy to be re-examined, which triggers clean filters
// and which has unexpected side effects (e.g. downloading filtered-out files)
type gitIndexer struct {
	cmd    *subprocess.Cmd
	input  io.WriteCloser
	output bytes.Buffer
	mu     sync.Mutex
}

func (i *gitIndexer) Add(path string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.cmd == nil {
		// Fire up the update-index command
		stdin, err := git.StartUpdateIndexFromStdin(&i.output)
		if err != nil {
			return err
		}

		i.input = stdin
	}

	i.input.Write([]byte(path + "\n"))
	return nil
}

func (i *gitIndexer) Output() string {
	return i.output.String()
}

func (i *gitIndexer) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.input != nil {
		i.input.Close()
	}

	if i.cmd != nil {
		return i.cmd.Wait()
	}

	return nil
}
