package commands

import (
	"github.com/github/git-media/filters"
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/metafile"
	"github.com/spf13/cobra"
	"os"
)

var (
	smudgeInfo = false
	smudgeCmd  = &cobra.Command{
		Use:   "smudge",
		Short: "Implements the Git smudge filter",
		Run:   smudgeCommand,
	}
)

func smudgeCommand(cmd *cobra.Command, args []string) {
	gitmedia.InstallHooks()

	pointer, err := metafile.Decode(os.Stdin)
	if err != nil {
		Panic(err, "Error reading git-media meta data from stdin:")
	}

	if smudgeInfo {
		localPath, err := gitmedia.LocalMediaPath(pointer.Oid)
		if err != nil {
			Exit(err.Error())
		}

		_, err = os.Stat(localPath)
		if err != nil {
			localPath = "--"
		}

		Print("%d %s", pointer.Size, localPath)
		return
	}

	err = filters.Smudge(os.Stdout, pointer.Oid)
	if err != nil {
		smudgerr := err.(*filters.SmudgeError)
		Panic(err, "Error reading file from local media dir: %s", smudgerr.Filename)
	}
}

func init() {
	smudgeCmd.Flags().BoolVarP(&smudgeInfo, "info", "i", false, "whatever")
	RootCmd.AddCommand(smudgeCmd)
}
