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
)

func pullCommand(cmd *cobra.Command, args []string) {

	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, "Could not pull")
	}

	c := fetchRefToChan(ref)
	checkoutAllFromFetchChan(c)
}

func init() {
	RootCmd.AddCommand(pullCmd)
}
