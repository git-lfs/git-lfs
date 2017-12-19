package commands

import (
	"os"
	"strings"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/tools/humanize"
	"github.com/spf13/cobra"
)

var (
	longOIDs        = false
	lsFilesShowSize = false
	debug           = false
)

func lsFilesCommand(cmd *cobra.Command, args []string) {
	requireInRepo()

	var ref string

	if len(args) == 1 {
		ref = args[0]
	} else {
		fullref, err := git.CurrentRef()
		if err != nil {
			ref = git.RefBeforeFirstCommit
		} else {
			ref = fullref.Sha
		}
	}

	showOidLen := 10
	if longOIDs {
		showOidLen = 64
	}

	seen := make(map[string]struct{})

	gitscanner := lfs.NewGitScanner(func(p *lfs.WrappedPointer, err error) {
		if _, ok := seen[p.Name]; ok {
			return
		}

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
				cfg.LFSObjectExists(p.Oid, p.Size),
				p.OidType,
				p.Oid,
				p.Version)
		} else {
			msg := []string{p.Oid[:showOidLen], lsFilesMarker(p), p.Name}
			if lsFilesShowSize {
				size := humanize.FormatBytes(uint64(p.Size))
				msg = append(msg, "("+size+")")
			}

			Print(strings.Join(msg, " "))
		}

		seen[p.Name] = struct{}{}
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
		cmd.Flags().BoolVarP(&lsFilesShowSize, "size", "s", false, "")
		cmd.Flags().BoolVarP(&debug, "debug", "d", false, "")
	})
}
