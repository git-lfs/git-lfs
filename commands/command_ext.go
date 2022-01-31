package commands

import (
	"fmt"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/spf13/cobra"
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

	for _, key := range args {
		ext := cfg.Extensions()[key]
		printExt(ext)
	}
}

func printAllExts() {
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
	Print(tr.Tr.Get("Extension: %s", ext.Name))
	Print(`    clean = %s
    smudge = %s
    priority = %d`, ext.Clean, ext.Smudge, ext.Priority)
}

func init() {
	RegisterCommand("ext", extCommand, func(cmd *cobra.Command) {
		cmd.AddCommand(NewCommand("list", extListCommand))
	})
}
