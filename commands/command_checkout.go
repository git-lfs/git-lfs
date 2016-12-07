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

	filter := filepathfilter.New(rootedpaths, nil)
	checkoutWithIncludeExclude(filter)
}

func checkoutFromFetchChan(in chan *lfs.WrappedPointer, filter *filepathfilter.Filter) {
	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not checkout")
	}

	// Need to ScanTree to identify multiple files with the same content (fetch will only report oids once)
	// use new gitscanner so mapping has all the scanned pointers before continuing
	mapping := make(map[string][]*lfs.WrappedPointer)
	chgitscanner := lfs.NewGitScanner(func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			Panic(err, "Could not scan for Git LFS files")
			return
		}
		mapping[p.Oid] = append(mapping[p.Oid], p)
	})
	chgitscanner.Filter = filter

	if err := chgitscanner.ScanTree(ref.Sha, nil); err != nil {
		ExitWithError(err)
	}

	chgitscanner.Close()

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

	if err := chgitscanner.ScanTree(ref.Sha, nil); err != nil {
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
