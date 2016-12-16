package commands

import (
	"fmt"

	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
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
	pull(filepathfilter.New(determineIncludeExcludePaths(cfg, includeArg, excludeArg)))

}

func pull(filter *filepathfilter.Filter) {
	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not pull")
	}

	c := fetchRefToChan(ref.Sha, filter)
	checkoutFromFetchChan(filter, c)
}

func init() {
	RegisterCommand("pull", pullCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&includeArg, "include", "I", "", "Include a list of paths")
		cmd.Flags().StringVarP(&excludeArg, "exclude", "X", "", "Exclude a list of paths")
	})
}
