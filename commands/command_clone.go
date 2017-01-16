package commands

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/localstorage"
	"github.com/git-lfs/git-lfs/subprocess"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/spf13/cobra"
)

var (
	cloneFlags git.CloneFlags
)

func cloneCommand(cmd *cobra.Command, args []string) {
	requireGitVersion()

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
	if err != nil || !tools.DirExists(clonedir) {
		// Derive from clone URL instead
		base := path.Base(args[len(args)-1])
		if strings.HasSuffix(base, ".git") {
			base = base[:len(base)-4]
		}
		clonedir, _ = filepath.Abs(base)
		if !tools.DirExists(clonedir) {
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
	localstorage.ResolveDirs()
	requireInRepo()

	// Now just call pull with default args
	// Support --origin option to clone
	var remote string
	if len(cloneFlags.Origin) > 0 {
		remote = cloneFlags.Origin
	} else {
		remote = "origin"
	}

	includeArg, excludeArg := getIncludeExcludeArgs(cmd)
	filter := buildFilepathFilter(cfg, includeArg, excludeArg)
	if cloneFlags.NoCheckout || cloneFlags.Bare {
		// If --no-checkout or --bare then we shouldn't check out, just fetch instead
		cfg.CurrentRemote = remote
		fetchRef("HEAD", filter)
	} else {
		pull(remote, filter)
		err := postCloneSubmodules(args)
		if err != nil {
			Exit("Error performing 'git lfs pull' for submodules: %v", err)
		}
	}
}

func postCloneSubmodules(args []string) error {
	// In git 2.9+ the filter option will have been passed through to submodules
	// So we need to lfs pull inside each
	if !git.Config.IsGitVersionAtLeast("2.9.0") {
		// In earlier versions submodules would have used smudge filter
		return nil
	}
	// Also we only do this if --recursive or --recurse-submodules was provided
	if !cloneFlags.Recursive && !cloneFlags.RecurseSubmodules {
		return nil
	}

	// Use `git submodule foreach --recursive` to cascade into nested submodules
	// Also good to call a new instance of git-lfs rather than do things
	// inside this instance, since that way we get a clean env in that subrepo
	cmd := subprocess.ExecCommand("git", "submodule", "foreach", "--recursive",
		"git lfs pull")
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func init() {
	RegisterCommand("clone", cloneCommand, func(cmd *cobra.Command) {
		cmd.PreRun = nil

		// Mirror all git clone flags
		cmd.Flags().StringVarP(&cloneFlags.TemplateDirectory, "template", "", "", "See 'git clone --help'")
		cmd.Flags().BoolVarP(&cloneFlags.Local, "local", "l", false, "See 'git clone --help'")
		cmd.Flags().BoolVarP(&cloneFlags.Shared, "shared", "s", false, "See 'git clone --help'")
		cmd.Flags().BoolVarP(&cloneFlags.NoHardlinks, "no-hardlinks", "", false, "See 'git clone --help'")
		cmd.Flags().BoolVarP(&cloneFlags.Quiet, "quiet", "q", false, "See 'git clone --help'")
		cmd.Flags().BoolVarP(&cloneFlags.NoCheckout, "no-checkout", "n", false, "See 'git clone --help'")
		cmd.Flags().BoolVarP(&cloneFlags.Progress, "progress", "", false, "See 'git clone --help'")
		cmd.Flags().BoolVarP(&cloneFlags.Bare, "bare", "", false, "See 'git clone --help'")
		cmd.Flags().BoolVarP(&cloneFlags.Mirror, "mirror", "", false, "See 'git clone --help'")
		cmd.Flags().StringVarP(&cloneFlags.Origin, "origin", "o", "", "See 'git clone --help'")
		cmd.Flags().StringVarP(&cloneFlags.Branch, "branch", "b", "", "See 'git clone --help'")
		cmd.Flags().StringVarP(&cloneFlags.Upload, "upload-pack", "u", "", "See 'git clone --help'")
		cmd.Flags().StringVarP(&cloneFlags.Reference, "reference", "", "", "See 'git clone --help'")
		cmd.Flags().BoolVarP(&cloneFlags.Dissociate, "dissociate", "", false, "See 'git clone --help'")
		cmd.Flags().StringVarP(&cloneFlags.SeparateGit, "separate-git-dir", "", "", "See 'git clone --help'")
		cmd.Flags().StringVarP(&cloneFlags.Depth, "depth", "", "", "See 'git clone --help'")
		cmd.Flags().BoolVarP(&cloneFlags.Recursive, "recursive", "", false, "See 'git clone --help'")
		cmd.Flags().BoolVarP(&cloneFlags.RecurseSubmodules, "recurse-submodules", "", false, "See 'git clone --help'")
		cmd.Flags().StringVarP(&cloneFlags.Config, "config", "c", "", "See 'git clone --help'")
		cmd.Flags().BoolVarP(&cloneFlags.SingleBranch, "single-branch", "", false, "See 'git clone --help'")
		cmd.Flags().BoolVarP(&cloneFlags.NoSingleBranch, "no-single-branch", "", false, "See 'git clone --help'")
		cmd.Flags().BoolVarP(&cloneFlags.Verbose, "verbose", "", false, "See 'git clone --help'")
		cmd.Flags().BoolVarP(&cloneFlags.Ipv4, "ipv4", "", false, "See 'git clone --help'")
		cmd.Flags().BoolVarP(&cloneFlags.Ipv6, "ipv6", "", false, "See 'git clone --help'")

		cmd.Flags().StringVarP(&includeArg, "include", "I", "", "Include a list of paths")
		cmd.Flags().StringVarP(&excludeArg, "exclude", "X", "", "Exclude a list of paths")
	})
}
