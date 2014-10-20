package commands

import (
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/scanner"
	"github.com/spf13/cobra"
)

var (
	lsFilesCmd = &cobra.Command{
		Use:   "ls-files",
		Short: "Show information about git media files",
		Run:   lsFilesCommand,
	}
)

func lsFilesCommand(cmd *cobra.Command, args []string) {
	var ref string
	var err error

	if len(args) == 1 {
		ref = args[0]
	} else {
		ref, err = gitmedia.CurrentRef()
		if err != nil {
			Panic(err, "Could not ls-files")
		}
	}

	pointers, err := scanner.Scan(ref, "")
	if err != nil {
		Panic(err, "Could not scan for git media files")
	}

	for _, p := range pointers {
		Print(p.Name)
	}
}

func init() {
	RootCmd.AddCommand(lsFilesCmd)
}
