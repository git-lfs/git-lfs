package commands

import (
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	lsFilesCmd = &cobra.Command{
		Use: "ls-files",
		Run: lsFilesCommand,
	}
)

func lsFilesCommand(cmd *cobra.Command, args []string) {
	requireInRepo()

	var ref string
	var err error

	if len(args) == 1 {
		ref = args[0]
	} else {
		fullref, err := git.CurrentRef()
		if err != nil {
			Exit(err.Error())
		}
		ref = fullref.Sha
	}

	scanOpt := &lfs.ScanRefsOptions{SkipDeletedBlobs: true}
	pointers, err := lfs.ScanRefs(ref, "", scanOpt)
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}

	for _, p := range pointers {
		Print(p.Name)
	}
}

func init() {
	RootCmd.AddCommand(lsFilesCmd)
}
