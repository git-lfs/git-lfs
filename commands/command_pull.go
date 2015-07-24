package commands

import (
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
	"sync"
)

var (
	pullCmd = &cobra.Command{
		Use:   "pull",
		Short: "Downloads LFS files for the current ref, and checks out",
		Run:   pullCommand,
	}
)

func pullCommand(cmd *cobra.Command, args []string) {

	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not pull")
	}

	pointers, err := lfs.ScanRefs(ref, "", nil)
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}

	fetchChan := make(chan *lfs.WrappedPointer)
	var wait sync.WaitGroup
	wait.Add(1)
	// Prep the checkout process for items that come out of fetch
	// this doesn't do anything except set up the goroutine & wait on the fetchChan
	checkoutWithChan(fetchChan, wait)
	// Do the downloading & report to channel which checkout watches
	fetchAndReportToChan(pointers, fetchChan)

	// Wait for final checkouts to finish
	wait.Wait()
}

func init() {
	RootCmd.AddCommand(pullCmd)
}
