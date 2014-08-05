package commands

import (
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/pointer"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
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
		pointer.Encode(os.Stdout, ptr)
		filename := smudgeFilename(args, err)
		Error("Error accessing media: %s (%s)", filename, ptr.Oid)
	}
}

func smudgeFilename(args []string, err error) string {
	if len(args) > 0 {
		return args[0]
	}

	if smudgeErr, ok := err.(*pointer.SmudgeError); ok {
		return filepath.Base(smudgeErr.Filename)
	}

	return "<unknown file>"
}

func init() {
	smudgeCmd.Flags().BoolVarP(&smudgeInfo, "info", "i", false, "whatever")
	RootCmd.AddCommand(smudgeCmd)
}
