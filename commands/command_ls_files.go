package commands

import (
	"strings"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	longOIDs   = false
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

	showShaLen := 7
	if longOIDs {
		showShaLen = 40
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
		Print(p.Sha1[0:showShaLen] + " ==> [ " + strings.Join(listFiles[p.Sha1], ",") + " ]")
		delete(listFiles, p.Sha1)
	}
}

func init() {
	lsFilesCmd.Flags().BoolVarP(&longOIDs, "long", "l", false, "Show object ID(s) 40 characters")
	RootCmd.AddCommand(lsFilesCmd)
}
