package commands

import (
	"github.com/github/git-lfs/git"
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

	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not pull")
	}

	includePaths, excludePaths := determineIncludeExcludePaths(pullIncludeArg, pullExcludeArg)

	c := fetchRefToChan(ref, includePaths, excludePaths)
	checkoutFromFetchChan(includePaths, excludePaths, c)
}

func init() {
	pullCmd.Flags().StringVarP(&pullIncludeArg, "include", "I", "", "Include a list of paths")
	pullCmd.Flags().StringVarP(&pullExcludeArg, "exclude", "X", "", "Exclude a list of paths")
	RootCmd.AddCommand(pullCmd)
}
