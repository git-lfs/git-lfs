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

func init() {
	RegisterCommand("pull", pullCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&includeArg, "include", "I", "", "Include a list of paths")
		cmd.Flags().StringVarP(&excludeArg, "exclude", "X", "", "Exclude a list of paths")
	})
}
