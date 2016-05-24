package commands

import (
	"github.com/github/git-lfs/httputil"
	"github.com/spf13/cobra"
)

var (
	lovesComics bool

	versionCmd = &cobra.Command{
		Use: "version",
		Run: versionCommand,
	}
)

func versionCommand(cmd *cobra.Command, args []string) {
	Print(httputil.UserAgent)

	if lovesComics {
		Print("Nothing may see Gah Lak Tus and survive!")
	}
}

func init() {
	versionCmd.Flags().BoolVarP(&lovesComics, "comics", "c", false, "easter egg")
	RootCmd.AddCommand(versionCmd)
}
