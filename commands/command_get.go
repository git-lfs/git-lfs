package commands

import (
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
		q.Add(p)
	}

	q.Process()
}

func init() {
	RootCmd.AddCommand(getCmd)
}
