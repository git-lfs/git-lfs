package commands

import (
	"os"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	longOIDs = false
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

	showOidLen := 10
	if longOIDs {
		showOidLen = 64
	}

	files, err := lfs.ScanTree(ref)
	if err != nil {
		Panic(err, "Could not scan for Git LFS tree: %s", err)
	}

	for _, p := range files {
		Print("%s %s %s", p.Oid[0:showOidLen], lsFilesMarker(p), p.Name)
	}
}

func lsFilesMarker(p *lfs.WrappedPointer) string {
	info, err := os.Stat(p.Name)
	if err == nil && info.Size() == p.Size {
		return "*"
	}

	return "-"
}

func init() {
	RegisterCommand("ls-files", lsFilesCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&longOIDs, "long", "l", false, "")
	})
}
