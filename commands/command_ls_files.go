package commands

import (
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	lsFilesCmd = &cobra.Command{
		Use:   "ls-files",
		Short: "Show information about Git LFS files",
		Run:   lsFilesCommand,
	}
)

func lsFilesCommand(cmd *cobra.Command, args []string) {
	var ref string
	var err error

	if len(args) == 1 {
		ref = args[0]
	} else {
		ref, err = git.CurrentRef()
		if err != nil {
			Panic(err, "Could not ls-files")
		}
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
