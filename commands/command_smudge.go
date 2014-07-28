package commands

import (
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/pointer"
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

	ptr, err := pointer.Decode(os.Stdin)
	if err != nil {
		Panic(err, "Error reading git-media meta data from stdin:")
	}

	if smudgeInfo {
		localPath, err := gitmedia.LocalMediaPath(ptr.Oid)
		if err != nil {
			Exit(err.Error())
		}

		stat, err := os.Stat(localPath)
		if err != nil {
			Print("%d --", ptr.Size)
		} else {
			Print("%d %s", stat.Size(), localPath)
		}
		return
	}

	err = pointer.Smudge(os.Stdout, ptr.Oid)
	if err != nil {
		smudgerr := err.(*pointer.SmudgeError)
		Panic(err, "Error reading file from local media dir: %s", smudgerr.Filename)
	}
}

func init() {
	smudgeCmd.Flags().BoolVarP(&smudgeInfo, "info", "i", false, "whatever")
	RootCmd.AddCommand(smudgeCmd)
}
