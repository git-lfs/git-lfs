package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/progress"
	"github.com/spf13/cobra"
)

func checkoutCommand(cmd *cobra.Command, args []string) {
	requireInRepo()
	filter := filepathfilter.New(rootedPaths(args), nil)
	checkoutWithIncludeExclude(filter)
}

func checkoutWithIncludeExclude(filter *filepathfilter.Filter) {
	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not checkout")
	}

	// this func has to load all pointers into memory
	var pointers []*lfs.WrappedPointer
	var multiErr error
	chgitscanner := lfs.NewGitScanner(func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			if multiErr != nil {
				multiErr = fmt.Errorf("%v\n%v", multiErr, err)
			} else {
				multiErr = err
			}
			return
		}

		pointers = append(pointers, p)
	})

	chgitscanner.Filter = filter

	if err := chgitscanner.ScanTree(ref.Sha); err != nil {
		ExitWithError(err)
	}
	chgitscanner.Close()

	if multiErr != nil {
		Panic(multiErr, "Could not scan for Git LFS files")
	}

	var wait sync.WaitGroup
	wait.Add(1)

	c := make(chan *lfs.WrappedPointer, 1)

	go func() {
		checkoutWithChan(c)
		wait.Done()
	}()

	meter := progress.NewMeter(progress.WithOSEnv(cfg.Os))
	meter.Start()
	var totalBytes int64
	for _, pointer := range pointers {
		totalBytes += pointer.Size
		meter.Add(totalBytes)
		meter.StartTransfer(pointer.Name)
		c <- pointer
		// not strictly correct (parallel) but we don't have a callback & it's just local
		// plus only 1 slot in channel so it'll block & be close
		meter.TransferBytes("checkout", pointer.Name, pointer.Size, totalBytes, int(pointer.Size))
		meter.FinishTransfer(pointer.Name)
	}
	close(c)
	wait.Wait()
	meter.Finish()
}

// Populate the working copy with the real content of objects where the file is
// either missing, or contains a matching pointer placeholder, from a list of pointers.
// If the file exists but has other content it is left alone
// Callers of this function MUST NOT Panic or otherwise exit the process
// without waiting for this function to shut down.  If the process exits while
// update-index is in the middle of processing a file the git index can be left
// in a locked state.
func checkoutWithChan(in <-chan *lfs.WrappedPointer) {
	// Get a converter from repo-relative to cwd-relative
	// Since writing data & calling git update-index must be relative to cwd
	pathConverter, err := lfs.NewRepoToCurrentPathConverter()
	if err != nil {
		Panic(err, "Could not convert file paths")
	}

	manifest := TransferManifest()
	gitIndexer := &gitIndexer{}

	// From this point on, git update-index is running. Code in this loop MUST
	// NOT Panic() or otherwise cause the process to exit. If the process exits
	// while update-index is in the middle of updating, the index can remain in a
	// locked state.

	// As files come in, write them to the wd and update the index
	for pointer := range in {
		// Check the content - either missing or still this pointer (not exist is ok)
		filepointer, err := lfs.DecodePointerFromFile(pointer.Name)
		if err != nil && !os.IsNotExist(err) {
			if errors.IsNotAPointerError(err) {
				// File has non-pointer content, leave it alone
				continue
			}
			LoggedError(err, "Problem accessing %v", pointer.Name)
			continue
		}

		if filepointer != nil && filepointer.Oid != pointer.Oid {
			// User has probably manually reset a file to another commit
			// while leaving it a pointer; don't mess with this
			continue
		}

		cwdfilepath := pathConverter.Convert(pointer.Name)

		err = lfs.PointerSmudgeToFile(cwdfilepath, pointer.Pointer, false, manifest, nil)
		if err != nil {
			if errors.IsDownloadDeclinedError(err) {
				// acceptable error, data not local (fetch not run or include/exclude)
				LoggedError(err, "Skipped checkout for %v, content not local. Use fetch to download.", pointer.Name)
			} else {
				LoggedError(err, "Could not checkout file")
				continue
			}
		}

		// errors are only returned when the gitIndexer is starting a new cmd
		if err := gitIndexer.Add(cwdfilepath); err != nil {
			Panic(err, "Could not update the index")
		}
	}

	if err := gitIndexer.Close(); err != nil {
		LoggedError(err, "Error updating the git index:\n%s", gitIndexer.Output())
	}
}

// Parameters are filters
// firstly convert any pathspecs to the root of the repo, in case this is being
// executed in a sub-folder
func rootedPaths(args []string) []string {
	pathConverter, err := lfs.NewCurrentToRepoPathConverter()
	if err != nil {
		Panic(err, "Could not checkout")
	}

	rootedpaths := make([]string, 0, len(args))
	for _, arg := range args {
		rootedpaths = append(rootedpaths, pathConverter.Convert(arg))
	}
	return rootedpaths
}

func init() {
	RegisterCommand("checkout", checkoutCommand, nil)
}

// Don't fire up the update-index command until we have at least one file to
// give it. Otherwise git interprets the lack of arguments to mean param-less update-index
// which can trigger entire working copy to be re-examined, which triggers clean filters
// and which has unexpected side effects (e.g. downloading filtered-out files)
type gitIndexer struct {
	cmd    *exec.Cmd
	input  io.WriteCloser
	output bytes.Buffer
	mu     sync.Mutex
}

func (i *gitIndexer) Add(path string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.cmd == nil {
		// Fire up the update-index command
		i.cmd = exec.Command("git", "update-index", "-q", "--refresh", "--stdin")
		i.cmd.Stdout = &i.output
		i.cmd.Stderr = &i.output
		stdin, err := i.cmd.StdinPipe()
		if err == nil {
			err = i.cmd.Start()
		}

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
