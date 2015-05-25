package commands

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
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
	requireStdin("This command should be run by the Git 'smudge' filter")
	lfs.InstallHooks(false)

	b := &bytes.Buffer{}
	r := io.TeeReader(os.Stdin, b)

	ptr, err := lfs.DecodePointer(r)
	if err != nil {
		mr := io.MultiReader(b, os.Stdin)
		_, err := io.Copy(os.Stdout, mr)
		if err != nil {
			Panic(err, "Error writing data to stdout:")
		}
		return
	}

	if smudgeInfo {
		localPath, err := lfs.LocalMediaPath(ptr.Oid)
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

	filename := smudgeFilename(args, err)
	cb, file, err := lfs.CopyCallbackFile("smudge", filename, 1, 1)
	if err != nil {
		Error(err.Error())
	}

	err = ptr.Smudge(os.Stdout, filename, cb)
	if file != nil {
		file.Close()
	}

	if err != nil {
		ptr.Encode(os.Stdout)
		LoggedError(err, "Error accessing media: %s (%s)", filename, ptr.Oid)
	}
}

func smudgeFilename(args []string, err error) string {
	if len(args) > 0 {
		return args[0]
	}

	if smudgeErr, ok := err.(*lfs.SmudgeError); ok {
		return filepath.Base(smudgeErr.Filename)
	}

	return "<unknown file>"
}

func init() {
	smudgeCmd.Flags().BoolVarP(&smudgeInfo, "info", "i", false, "Display the local path and size of the smudged file.")
	RootCmd.AddCommand(smudgeCmd)
}
