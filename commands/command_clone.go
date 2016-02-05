package commands

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	cloneCmd = &cobra.Command{
		Use: "clone",
		Run: cloneCommand,
	}
)

func cloneCommand(cmd *cobra.Command, args []string) {

	// We pass all args to git clone
	err := git.CloneWithoutFilters(args)
	if err != nil {
		Panic(err, "Error(s) during clone")
	}

	// now execute pull (need to be inside dir)
	cwd, err := os.Getwd()
	if err != nil {
		Panic(err, "Unable to derive current working dir")
	}

	// Either the last argument was a relative or local dir, or we have to
	// derive it from the clone URL
	clonedir, err := filepath.Abs(args[len(args)-1])
	if err != nil || !lfs.DirExists(clonedir) {
		// Derive from clone URL instead
		base := path.Base(args[len(args)-1])
		if strings.HasSuffix(base, ".git") {
			base = base[:len(base)-4]
		}
		clonedir, _ = filepath.Abs(base)
		if !lfs.DirExists(clonedir) {
			Exit("Unable to find clone dir at %q", clonedir)
		}
	}

	err = os.Chdir(clonedir)
	if err != nil {
		Panic(err, "Unable to change directory to clone dir %q", clonedir)
	}

	// Make sure we pop back to dir we started in at the end
	defer os.Chdir(cwd)

	// Also need to derive dirs now
	lfs.ResolveDirs()
	requireInRepo()

	// Now just call pull with default args
	lfs.Config.CurrentRemote = "origin" // always origin after clone
	pull(nil, nil)

}

func init() {
	RootCmd.AddCommand(cloneCmd)
}
