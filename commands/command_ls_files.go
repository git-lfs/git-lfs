package commands

import (
	"os"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
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

	gitscanner := lfs.NewGitScanner(nil)
	defer gitscanner.Close()

	pointerCh, err := gitscanner.ScanTree(ref)
	if err != nil {
		Exit("Could not scan for Git LFS tree: %s", err)
	}

	for p := range pointerCh.Results {
		Print("%s %s %s", p.Oid[0:showOidLen], lsFilesMarker(p), p.Name)
	}

	if err := pointerCh.Wait(); err != nil {
		Exit("Could not scan for Git LFS tree: %s", err)
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
