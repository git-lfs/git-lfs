package commands

import (
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	pullCmd = &cobra.Command{
		Use:   "pull",
		Short: "Downloads LFS files for the current ref, and checks out",
		Run:   pullCommand,
	}
	pullIncludeArg string
	pullExcludeArg string
)

func pullCommand(cmd *cobra.Command, args []string) {
	requireInRepo()

	if len(args) > 0 {
		// Remote is first arg
		lfs.Config.CurrentRemote = args[0]
	} else {
		trackedRemote, err := git.CurrentRemote()
		if err == nil {
			lfs.Config.CurrentRemote = trackedRemote
		} // otherwise leave as default (origin)
	}

	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not pull")
	}

	includePaths, excludePaths := determineIncludeExcludePaths(pullIncludeArg, pullExcludeArg)

	c := fetchRefToChan(ref.Sha, includePaths, excludePaths)
	checkoutFromFetchChan(includePaths, excludePaths, c)
}

func init() {
	pullCmd.Flags().StringVarP(&pullIncludeArg, "include", "I", "", "Include a list of paths")
	pullCmd.Flags().StringVarP(&pullExcludeArg, "exclude", "X", "", "Exclude a list of paths")
	RootCmd.AddCommand(pullCmd)
}
