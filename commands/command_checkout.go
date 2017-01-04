package commands

import (
	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/progress"
	"github.com/spf13/cobra"
)

func checkoutCommand(cmd *cobra.Command, args []string) {
	requireInRepo()
	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not checkout")
	}

	var totalBytes int64
	meter := progress.NewMeter(progress.WithOSEnv(cfg.Os))
	singleCheckout := newSingleCheckout("")
	chgitscanner := lfs.NewGitScanner(func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			LoggedError(err, "Scanner error")
			return
		}

		totalBytes += p.Size
		meter.Add(p.Size)
		meter.StartTransfer(p.Name)

		singleCheckout.Run(p)

		// not strictly correct (parallel) but we don't have a callback & it's just local
		// plus only 1 slot in channel so it'll block & be close
		meter.TransferBytes("checkout", p.Name, p.Size, totalBytes, int(p.Size))
		meter.FinishTransfer(p.Name)
	})

	chgitscanner.Filter = filepathfilter.New(rootedPaths(args), nil)

	if err := chgitscanner.ScanTree(ref.Sha); err != nil {
		ExitWithError(err)
	}

	meter.Start()
	chgitscanner.Close()
	meter.Finish()
	singleCheckout.Close()
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
