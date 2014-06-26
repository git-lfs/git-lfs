package commands

import (
	"github.com/github/git-media/gitmedia"
	"github.com/spf13/cobra"
)

var (
	lovesComics bool

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the version number.",
		Run:   versionCommand,
	}
)

func versionCommand(cmd *cobra.Command, args []string) {
	var parent *cobra.Command
	if p := cmd.Parent(); p != nil {
		parent = p
	} else {
		parent = cmd
	}

	Print("%s v%s", parent.Name(), gitmedia.Version)

	if lovesComics {
		Print("Nothing may see Gah Lak Tus and survive!")
	}
}

func init() {
	versionCmd.Flags().BoolVarP(&lovesComics, "comics", "c", false, "easter egg")
	RootCmd.AddCommand(versionCmd)
}
