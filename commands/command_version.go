package commands

import (
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/spf13/cobra"
)

var (
	lovesComics bool
)

func versionCommand(cmd *cobra.Command, args []string) {
	Print(lfsapi.UserAgent)

	if lovesComics {
		Print("Nothing may see Gah Lak Tus and survive!")
	}
}

func init() {
	RegisterCommand("version", versionCommand, func(cmd *cobra.Command) {
		cmd.PreRun = nil
		cmd.Flags().BoolVarP(&lovesComics, "comics", "c", false, "easter egg")
	})
}
