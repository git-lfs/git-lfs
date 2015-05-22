package commands

import (
	"os"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	getCmd = &cobra.Command{
		Use:   "get",
		Short: "get",
		Run:   getCommand,
	}
)

func getCommand(cmd *cobra.Command, args []string) {
	var ref string
	var err error

	if len(args) == 1 {
		ref = args[0]
	} else {
		ref, err = git.CurrentRef()
		if err != nil {
			Panic(err, "Could not get")
		}
	}

	pointers, err := lfs.ScanRefs(ref, "")
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}

	q := lfs.NewDownloadQueue(lfs.Config.ConcurrentTransfers(), len(pointers))

	for _, p := range pointers {
		q.Add(lfs.NewDownloadable(p))
	}

	q.Process()

	target, err := git.ResolveRef(ref)
	if err != nil {
	}

	current, err := git.CurrentRef()
	if err != nil {
	}

	if target == current {
		// We just downloaded the files for the current ref, we can copy them into
		// the working directory and update the git index
		for _, pointer := range pointers {
			file, err := os.Create(pointer.Name)
			if err != nil {
				Panic(err, "Could not create working directory file")
			}

			if err := lfs.PointerSmudge(file, pointer.Pointer, pointer.Name, nil); err != nil {
				Panic(err, "Could not write working directory file")
			}

			if err := git.UpdateIndex(pointer.Name); err != nil {
				Panic(err, "Could not update index")
			}
		}
	}
}

func init() {
	RootCmd.AddCommand(getCmd)
}
