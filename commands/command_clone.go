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

	cloneFlags git.CloneFlags
)

func cloneCommand(cmd *cobra.Command, args []string) {

	// We pass all args to git clone
	err := git.CloneWithoutFilters(cloneFlags, args)
	if err != nil {
		Exit("Error(s) during clone:\n%v", err)
	}

	// now execute pull (need to be inside dir)
	cwd, err := os.Getwd()
	if err != nil {
		Exit("Unable to derive current working dir: %v", err)
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
		Exit("Unable to change directory to clone dir %q: %v", clonedir, err)
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
	// Mirror all git clone flags
	cloneCmd.Flags().StringVarP(&cloneFlags.TemplateDirectory, "template", "", "", "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Local, "local", "l", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Shared, "shared", "s", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.NoHardlinks, "no-hardlinks", "", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Quiet, "quiet", "q", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.NoCheckout, "no-checkout", "n", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Progress, "progress", "", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Bare, "bare", "", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Mirror, "mirror", "", false, "See 'git clone --help'")
	cloneCmd.Flags().StringVarP(&cloneFlags.Origin, "origin", "o", "", "See 'git clone --help'")
	cloneCmd.Flags().StringVarP(&cloneFlags.Branch, "branch", "b", "", "See 'git clone --help'")
	cloneCmd.Flags().StringVarP(&cloneFlags.Upload, "upload-pack", "u", "", "See 'git clone --help'")
	cloneCmd.Flags().StringVarP(&cloneFlags.Reference, "reference", "", "", "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Dissociate, "dissociate", "", false, "See 'git clone --help'")
	cloneCmd.Flags().StringVarP(&cloneFlags.SeparateGit, "separate-git-dir", "", "", "See 'git clone --help'")
	cloneCmd.Flags().StringVarP(&cloneFlags.Depth, "depth", "", "", "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Recursive, "recursive", "", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.RecurseSubmodules, "recurse-submodules", "", false, "See 'git clone --help'")
	cloneCmd.Flags().StringVarP(&cloneFlags.Config, "config", "c", "", "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.SingleBranch, "single-branch", "", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.NoSingleBranch, "no-single-branch", "", false, "See 'git clone --help'")
	RootCmd.AddCommand(cloneCmd)
}
