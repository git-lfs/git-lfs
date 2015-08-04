package commands

import (
	"fmt"

	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	extCmd = &cobra.Command{
		Use:   "ext",
		Short: "View details for all extensions",
		Run:   extCommand,
	}

	extListCmd = &cobra.Command{
		Use:   "list",
		Short: "View details for specified extensions",
		Run:   extListCommand,
	}
)

func extCommand(cmd *cobra.Command, args []string) {
	printAllExts()
}

func extListCommand(cmd *cobra.Command, args []string) {
	n := len(args)
	if n == 0 {
		printAllExts()
		return
	}

	config := lfs.Config
	for _, key := range args {
		ext := config.Extensions()[key]
		printExt(ext)
	}
}

func printAllExts() {
	config := lfs.Config

	extensions, err := lfs.SortExtensions(config.Extensions())
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, ext := range extensions {
		printExt(ext)
	}
}

func printExt(ext lfs.Extension) {
	Print("Extension: %s", ext.Name)
	Print("    clean = %s", ext.Clean)
	Print("    smudge = %s", ext.Smudge)
	Print("    priority = %d", ext.Priority)
}

func init() {
	extCmd.AddCommand(extListCmd)
	RootCmd.AddCommand(extCmd)
}
