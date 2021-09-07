package commands

import (
	"github.com/git-lfs/git-lfs/v3/lfshttp"
	"github.com/spf13/cobra"
)

var (
	lovesComics bool
)

func versionCommand(cmd *cobra.Command, args []string) {
	Print(lfshttp.UserAgent)

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
