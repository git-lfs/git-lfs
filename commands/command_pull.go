package commands

import (
	"fmt"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/spf13/cobra"
)

func pullCommand(cmd *cobra.Command, args []string) {
	requireGitVersion()
	requireInRepo()

	if len(args) > 0 {
		// Remote is first arg
		if err := git.ValidateRemote(args[0]); err != nil {
			Panic(err, fmt.Sprintf("Invalid remote name '%v'", args[0]))
		}
		cfg.CurrentRemote = args[0]
	} else {
		// Actively find the default remote, don't just assume origin
		defaultRemote, err := git.DefaultRemote()
		if err != nil {
			Panic(err, "No default remote")
		}
		cfg.CurrentRemote = defaultRemote
	}

	includeArg, excludeArg := getIncludeExcludeArgs(cmd)
	include, exclude := determineIncludeExcludePaths(cfg, includeArg, excludeArg)
	gitscanner := lfs.NewGitScanner()
	defer gitscanner.Close()
	pull(gitscanner, include, exclude)
}

func pull(gitscanner *lfs.GitScanner, includePaths, excludePaths []string) {
	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not pull")
	}

	c := fetchRefToChan(gitscanner, ref.Sha, includePaths, excludePaths)
	checkoutFromFetchChan(gitscanner, includePaths, excludePaths, c)
}

func init() {
	RegisterCommand("pull", pullCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&includeArg, "include", "I", "", "Include a list of paths")
		cmd.Flags().StringVarP(&excludeArg, "exclude", "X", "", "Exclude a list of paths")
	})
}
