package commands

import (
	"github.com/github/git-lfs/httputil"
	"github.com/spf13/cobra"
)

var (
	lovesComics bool
)

func versionCommand(cmd *cobra.Command, args []string) {
	Print(httputil.UserAgent)

	if lovesComics {
		Print("Nothing may see Gah Lak Tus and survive!")
	}
}

func init() {
	RegisterSubcommand(func() *cobra.Command {
		cmd := &cobra.Command{
			Use:    "version",
			PreRun: resolveLocalStorage,
			Run:    versionCommand,
		}

		cmd.Flags().BoolVarP(&lovesComics, "comics", "c", false, "easter egg")
		return cmd
	})
}
