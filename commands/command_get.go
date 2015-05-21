package commands

import (
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
	"os"
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
		for _, pointer := range pointers {
			file, err := os.Create(pointer.Name)
			if err != nil {
				Panic(err, "Could not create working directory file")
			}

			err = lfs.PointerSmudge(file, pointer.Pointer, pointer.Name, nil)
			if err != nil {
				Panic(err, "Could not write working directory file")
			}
		}
	}
}

func init() {
	RootCmd.AddCommand(getCmd)
}
