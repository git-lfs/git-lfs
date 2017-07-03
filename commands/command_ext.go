package commands

import (
	"fmt"

	"github.com/git-lfs/git-lfs/config"
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
	Print("Extension: %s", ext.Name)
	Print("    clean = %s", ext.Clean)
	Print("    smudge = %s", ext.Smudge)
	Print("    priority = %d", ext.Priority)
}

func init() {
	RegisterCommand("ext", extCommand, func(cmd *cobra.Command) {
		cmd.AddCommand(NewCommand("list", extListCommand))
	})
}
