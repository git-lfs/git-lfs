package commands

import (
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
	"os"
	"os/exec"
	"sync"
)

var (
	checkoutCmd = &cobra.Command{
		Use:   "checkout",
		Short: "checkout",
		Run:   checkoutCommand,
	}
)

func checkoutCommand(cmd *cobra.Command, args []string) {
	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not checkout")
	}

	pointers, err := lfs.ScanRefs(ref, "", nil)
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}

	var wait sync.WaitGroup
	wait.Add(1)

	c := make(chan *lfs.WrappedPointer)

	checkoutWithChan(c, wait)
	for _, pointer := range pointers {
		c <- pointer
	}
	close(c)
	wait.Wait()
}

func init() {
	RootCmd.AddCommand(checkoutCmd)
}

// Populate the working copy with the real content of objects where the file is
// either missing, or contains a matching pointer placeholder, from a list of pointers.
// If the file exists but has other content it is left alone
// returns immediately but a goroutine listens on the in channel for objects
// calls wait.Done() when the final item after the channel is closed is done
func checkoutWithChan(in <-chan *lfs.WrappedPointer, wait sync.WaitGroup) {
	go func() {
		// Fire up the update-index command
		cmd := exec.Command("git", "update-index", "-q", "--refresh", "--stdin")
		updateIdxStdin, err := cmd.StdinPipe()
		if err != nil {
			Panic(err, "Could not update the index")
		}

		if err := cmd.Start(); err != nil {
			Panic(err, "Could not update the index")
		}

		// As files come in, write them to the wd and update the index
		for pointer := range in {

			// Check the content - either missing or still this pointer (not exist is ok)
			filepointer, err := lfs.DecodePointerFromFile(pointer.Name)
			if err != nil && !os.IsNotExist(err) {
				if err == lfs.NotAPointerError {
					// File has non-pointer content, leave it alone
					continue
				}
				Panic(err, "Problem accessing %v", pointer.Name)
			}
			if filepointer != nil && filepointer.Oid != pointer.Oid {
				// User has probably manually reset a file to another commit
				// while leaving it a pointer; don't mess with this
				continue
			}
			// OK now we can (over)write the file content
			file, err := os.Create(pointer.Name)
			if err != nil {
				Panic(err, "Could not create working directory file")
			}

			if err := lfs.PointerSmudge(file, pointer.Pointer, pointer.Name, nil); err != nil {
				Panic(err, "Could not write working directory file")
			}
			file.Close()

			updateIdxStdin.Write([]byte(pointer.Name + "\n"))
		}

		updateIdxStdin.Close()
		if err := cmd.Wait(); err != nil {
			Panic(err, "Error updating the git index")
		}
		wait.Done()
	}()

}
