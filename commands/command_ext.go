package commands

import (
	"fmt"

	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	extCmd = &cobra.Command{
		Use: "ext",
		Run: extCommand,
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

	cfg := config.Config
	for _, key := range args {
		ext := cfg.Extensions()[key]
		printExt(ext)
	}
}

func printAllExts() {
	cfg := config.Config

	extensions, err := cfg.SortedExtensions()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, ext := range extensions {
		printExt(ext)
	}
}

func printExt(ext config.Extension) {
	Print("Extension: %s", ext.Name)
	Print("    clean = %s", ext.Clean)
	Print("    smudge = %s", ext.Smudge)
	Print("    priority = %d", ext.Priority)
}

func init() {
	extCmd.AddCommand(extListCmd)
	RootCmd.AddCommand(extCmd)
}
