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

	// Previously we would only checkout files that were downloaded, as they
	// were downloaded. However this would ignore files where the content was
	// already present locally (since these are no longer included in transfer Q for
	// better reporting purposes).
	// So now we do exactly what we say on the tin, fetch then a separate checkout
	fetchRef(ref)
	checkoutAll()
}

func init() {
	RootCmd.AddCommand(pullCmd)
}
