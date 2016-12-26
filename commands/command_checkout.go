package commands

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/progress"
	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
)

func checkoutCommand(cmd *cobra.Command, args []string) {
	requireInRepo()

	// Parameters are filters
	// firstly convert any pathspecs to the root of the repo, in case this is being executed in a sub-folder
	var rootedpaths []string

	inchan := make(chan string, 1)
	outchan, err := lfs.ConvertCwdFilesRelativeToRepo(inchan)
	if err != nil {
		Panic(err, "Could not checkout")
	}
	for _, arg := range args {
		inchan <- arg
		rootedpaths = append(rootedpaths, <-outchan)
	}
	close(inchan)
	checkoutWithIncludeExclude(filepathfilter.New(rootedpaths, nil))
}

// Checkout from items reported from the fetch process (in parallel)
func checkoutAllFromFetchChan(c chan *lfs.WrappedPointer) {
	tracerx.Printf("starting fetch/parallel checkout")
	checkoutFromFetchChan(nil, c)
}

func checkoutFromFetchChan(filter *filepathfilter.Filter, in chan *lfs.WrappedPointer) {
	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not checkout")
	}
	// Need to ScanTree to identify multiple files with the same content (fetch will only report oids once)
	pointers, err := lfs.ScanTree(ref.Sha)
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}

	// Map oid to multiple pointers
	mapping := make(map[string][]*lfs.WrappedPointer)
	for _, pointer := range pointers {
		if filter.Allows(pointer.Name) {
			mapping[pointer.Oid] = append(mapping[pointer.Oid], pointer)
		}
	}

	// Launch git update-index
	c := make(chan *lfs.WrappedPointer)

	var wait sync.WaitGroup
	wait.Add(1)

	go func() {
		checkoutWithChan(c)
		wait.Done()
	}()

	// Feed it from in, which comes from fetch
	for p := range in {
		// Add all of the files for this oid
		for _, fp := range mapping[p.Oid] {
			c <- fp
		}
	}
	close(c)
	wait.Wait()
}

func checkoutWithIncludeExclude(filter *filepathfilter.Filter) {
	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not checkout")
	}

	pointers, err := lfs.ScanTree(ref.Sha)
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}

	var wait sync.WaitGroup
	wait.Add(1)

	c := make(chan *lfs.WrappedPointer, 1)

	go func() {
		checkoutWithChan(c)
		wait.Done()
	}()

	// Count bytes for progress
	var totalBytes int64
	for _, pointer := range pointers {
		totalBytes += pointer.Size
	}

	logPath, _ := cfg.Os.Get("GIT_LFS_PROGRESS")
	progress := progress.NewProgressMeter(len(pointers), totalBytes, false, logPath)
	progress.Start()
	totalBytes = 0
	for _, pointer := range pointers {
		totalBytes += pointer.Size
		if filter.Allows(pointer.Name) {
			progress.Add(pointer.Name)
			c <- pointer
			// not strictly correct (parallel) but we don't have a callback & it's just local
			// plus only 1 slot in channel so it'll block & be close
			progress.TransferBytes("checkout", pointer.Name, pointer.Size, totalBytes, int(pointer.Size))
			progress.FinishTransfer(pointer.Name)
		} else {
			progress.Skip(pointer.Size)
		}
	}
	close(c)
	wait.Wait()
	progress.Finish()

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
	repopathchan := make(chan string, 1)
	cwdpathchan, err := lfs.ConvertRepoFilesRelativeToCwd(repopathchan)
	if err != nil {
		Panic(err, "Could not convert file paths")
	}

	// Don't fire up the update-index command until we have at least one file to
	// give it. Otherwise git interprets the lack of arguments to mean param-less update-index
	// which can trigger entire working copy to be re-examined, which triggers clean filters
	// and which has unexpected side effects (e.g. downloading filtered-out files)
	var cmd *exec.Cmd
	var updateIdxStdin io.WriteCloser
	var updateIdxOut bytes.Buffer

	// From this point on, git update-index is running. Code in this loop MUST
	// NOT Panic() or otherwise cause the process to exit. If the process exits
	// while update-index is in the middle of updating, the index can remain in a
	// locked state.

	// As files come in, write them to the wd and update the index

	manifest := TransferManifest()

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

		repopathchan <- pointer.Name
		cwdfilepath := <-cwdpathchan

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

		if cmd == nil {
			// Fire up the update-index command
			cmd = exec.Command("git", "update-index", "-q", "--refresh", "--stdin")
			cmd.Stdout = &updateIdxOut
			cmd.Stderr = &updateIdxOut
			updateIdxStdin, err = cmd.StdinPipe()
			if err != nil {
				Panic(err, "Could not update the index")
			}

			if err := cmd.Start(); err != nil {
				Panic(err, "Could not update the index")
			}

		}

		updateIdxStdin.Write([]byte(cwdfilepath + "\n"))
	}
	close(repopathchan)

	if cmd != nil && updateIdxStdin != nil {
		updateIdxStdin.Close()
		if err := cmd.Wait(); err != nil {
			LoggedError(err, "Error updating the git index:\n%s", updateIdxOut.String())
		}
	}
}

func init() {
	RegisterCommand("checkout", checkoutCommand, nil)
}
