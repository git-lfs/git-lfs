package commands

import (
	"fmt"
	"os"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/progress"
	"github.com/git-lfs/git-lfs/transfer"
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
	filter := filepathfilter.New(rootedPaths(args), nil)
	singleCheckout := newSingleCheckout()
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

	chgitscanner.Filter = filter

	if err := chgitscanner.ScanTree(ref.Sha); err != nil {
		ExitWithError(err)
	}

	meter.Start()
	chgitscanner.Close()
	meter.Finish()
	singleCheckout.Close()
}

func checkout(pointer *lfs.WrappedPointer, pathConverter lfs.PathConverter, manifest *transfer.Manifest) (string, error) {
	// Check the content - either missing or still this pointer (not exist is ok)
	filepointer, err := lfs.DecodePointerFromFile(pointer.Name)
	if err != nil && !os.IsNotExist(err) {
		if errors.IsNotAPointerError(err) {
			// File has non-pointer content, leave it alone
			return "", nil
		}
		return "", err
	}

	if filepointer != nil && filepointer.Oid != pointer.Oid {
		// User has probably manually reset a file to another commit
		// while leaving it a pointer; don't mess with this
		return "", nil
	}

	cwdfilepath := pathConverter.Convert(pointer.Name)

	err = lfs.PointerSmudgeToFile(cwdfilepath, pointer.Pointer, false, manifest, nil)
	if err != nil {
		if errors.IsDownloadDeclinedError(err) {
			// acceptable error, data not local (fetch not run or include/exclude)
			return "", fmt.Errorf("Skipped checkout for %q, content not local. Use fetch to download.", pointer.Name)
		} else {
			return "", fmt.Errorf("Could not check out %q", pointer.Name)
		}
	}

	return cwdfilepath, nil
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
