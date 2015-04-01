package commands

import (
	"github.com/hawser/git-hawser/hawser"
	"github.com/spf13/cobra"
)

var (
	pathCmd = &cobra.Command{
		Use:   "path",
		Short: "Manipulate .gitattributes",
		Run:   pathCommand,
	}
)

func pathCommand(cmd *cobra.Command, args []string) {
	hawser.InstallHooks(false)

	Print("Listing paths")
	knownPaths := findPaths()
	for _, t := range knownPaths {
		Print("    %s (%s)", t.Path, t.Source)
	}
}

func init() {
	RootCmd.AddCommand(pathCmd)
}
