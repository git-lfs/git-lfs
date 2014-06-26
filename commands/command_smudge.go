package commands

import (
	"github.com/github/git-media/filters"
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/metafile"
	"github.com/spf13/cobra"
	"os"
)

var (
	smudgeCmd = &cobra.Command{
		Use:   "smudge",
		Short: "Implements the Git smudge filter",
		Run:   smudgeCommand,
	}
)

func smudgeCommand(cmd *cobra.Command, args []string) {
	gitmedia.InstallHooks()

	sha, err := metafile.Decode(os.Stdin)
	if err != nil {
		Panic(err, "Error reading git-media meta data from stdin:")
	}

	err = filters.Smudge(os.Stdout, sha)
	if err != nil {
		smudgerr := err.(*filters.SmudgeError)
		Panic(err, "Error reading file from local media dir: %s", smudgerr.Filename)
	}
}

func init() {
	RootCmd.AddCommand(smudgeCmd)
}
