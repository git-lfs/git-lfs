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
	listFiles := make(map[string][]string)
	fileTree, err := lfs.ScanTree(ref)
	if err != nil {
		Panic(err, "Could not scan for Git LFS tree")
	}
	for _, p := range fileTree {
		listFiles[p.Sha1] = append(listFiles[p.Sha1], p.Name)
	}

	pointers, err := lfs.ScanRefs(ref, "", scanOpt)
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}
	for _, p := range pointers {
		Print(p.Name)
		if len(listFiles[p.Sha1]) > 1 {
			for _, v := range listFiles[p.Sha1][1:] {
				Print(v + " (duplicate of " + p.Name + ")")
			}
		}
		delete(listFiles, p.Sha1)
	}
}

func init() {
	RootCmd.AddCommand(lsFilesCmd)
}
