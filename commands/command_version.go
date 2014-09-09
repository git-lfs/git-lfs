package commands

import (
	"github.com/github/git-media/gitconfig"
	"github.com/github/git-media/gitmedia"
	"github.com/spf13/cobra"
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
	Print(gitmedia.UserAgent)

	v, err := gitconfig.Version()
	if err != nil {
		Print("Error getting git version: %s", err.Error())
	} else {
		Print(v)
	}

	if lovesComics {
		Print("Nothing may see Gah Lak Tus and survive!")
	}
}

func init() {
	versionCmd.Flags().BoolVarP(&lovesComics, "comics", "c", false, "easter egg")
	RootCmd.AddCommand(versionCmd)
}
