package commands

import (
	"os"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	longOIDs = false
	debug    = false
)

func lsFilesCommand(cmd *cobra.Command, args []string) {
	requireInRepo()

	var ref string

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

	gitscanner := lfs.NewGitScanner(func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			Exit("Could not scan for Git LFS tree: %s", err)
			return
		}

		if debug {
			Print(
				"filepath: %s\n"+
					"    size: %d\n"+
					"checkout: %v\n"+
					"download: %v\n"+
					"     oid: %s %s\n"+
					" version: %s\n",
				p.Name,
				p.Size,
				fileExistsOfSize(p),
				lfs.ObjectExistsOfSize(p.Oid, p.Size),
				p.OidType,
				p.Oid,
				p.Version)
		} else {
			Print("%s %s %s", p.Oid[0:showOidLen], lsFilesMarker(p), p.Name)
		}
	})
	defer gitscanner.Close()

	if err := gitscanner.ScanTree(ref); err != nil {
		Exit("Could not scan for Git LFS tree: %s", err)
	}
}

// Returns true if a pointer appears to be properly smudge on checkout
func fileExistsOfSize(p *lfs.WrappedPointer) bool {
	info, err := os.Stat(p.Name)
	return err == nil && info.Size() == p.Size
}

func lsFilesMarker(p *lfs.WrappedPointer) string {
	if fileExistsOfSize(p) {
		return "*"
	}
	return "-"
}

func init() {
	RegisterCommand("ls-files", lsFilesCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&longOIDs, "long", "l", false, "")
		cmd.Flags().BoolVarP(&debug, "debug", "d", false, "")
	})
}
