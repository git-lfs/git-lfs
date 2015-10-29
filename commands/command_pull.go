package commands

import (
	"fmt"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	pullCmd = &cobra.Command{
		Use: "pull",
		Run: pullCommand,
	}
	pullIncludeArg string
	pullExcludeArg string
)

func pullCommand(cmd *cobra.Command, args []string) {
	requireInRepo()

	if len(args) > 0 {
		// Remote is first arg
		if err := git.ValidateRemote(args[0]); err != nil {
			Panic(err, fmt.Sprintf("Invalid remote name '%v'", args[0]))
		}
		lfs.Config.CurrentRemote = args[0]
	} else {
		// Actively find the default remote, don't just assume origin
		defaultRemote, err := git.DefaultRemote()
		if err != nil {
			Panic(err, "No default remote")
		}
		lfs.Config.CurrentRemote = defaultRemote
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
