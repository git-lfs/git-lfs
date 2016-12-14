package commands

import (
	"fmt"
	"sync"

	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/spf13/cobra"
)

func pullCommand(cmd *cobra.Command, args []string) {
	requireGitVersion()
	requireInRepo()

	if len(args) > 0 {
		// Remote is first arg
		if err := git.ValidateRemote(args[0]); err != nil {
			Panic(err, fmt.Sprintf("Invalid remote name '%v'", args[0]))
		}
		cfg.CurrentRemote = args[0]
	} else {
		// Actively find the default remote, don't just assume origin
		defaultRemote, err := git.DefaultRemote()
		if err != nil {
			Panic(err, "No default remote")
		}
		cfg.CurrentRemote = defaultRemote
	}

	includeArg, excludeArg := getIncludeExcludeArgs(cmd)
	filter := buildFilepathFilter(cfg, includeArg, excludeArg)
	pull(filter)
}

func pull(filter *filepathfilter.Filter) {
	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not pull")
	}

	c := fetchRefToChan(ref.Sha, filter)
	checkoutFromFetchChan(c, filter)
}

func fetchRefToChan(ref string, filter *filepathfilter.Filter) chan *lfs.WrappedPointer {
	c := make(chan *lfs.WrappedPointer)
	pointers, err := pointersToFetchForRef(ref, filter)
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}

	go fetchAndReportToChan(pointers, filter, c)

	return c
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

	if err := chgitscanner.ScanTree(ref.Sha); err != nil {
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
		cwdfilepath, err := checkout(pointer, pathConverter, manifest)
		if err != nil {
			LoggedError(err, "Checkout error")
		}

		if len(cwdfilepath) == 0 {
			continue
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

func init() {
	RegisterCommand("pull", pullCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&includeArg, "include", "I", "", "Include a list of paths")
		cmd.Flags().StringVarP(&excludeArg, "exclude", "X", "", "Exclude a list of paths")
	})
}
