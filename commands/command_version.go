package commands

import (
	"github.com/github/git-lfs/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/github/git-lfs/lfs"
)

var (
	lovesComics bool

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the version number",
		Run:   versionCommand,
	}
)

func versionCommand(cmd *cobra.Command, args []string) {
	Print(lfs.UserAgent)

	if lovesComics {
		Print("Nothing may see Gah Lak Tus and survive!")
	}
}

func init() {
	versionCmd.Flags().BoolVarP(&lovesComics, "comics", "c", false, "easter egg")
	RootCmd.AddCommand(versionCmd)
}
