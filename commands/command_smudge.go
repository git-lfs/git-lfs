package commands

import (
	"fmt"
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

	filename := smudgeFilename(args, err)
	var cb gitmedia.CopyCallback
	var file *os.File

	if cbFile := os.Getenv("GIT_MEDIA_PROGRESS"); len(cbFile) > 0 {
		cbDir := filepath.Dir(cbFile)
		if err = os.MkdirAll(cbDir, 0755); err != nil {
			Error("Error writing Git Media progress to %s: %s", cbFile, err.Error())
		}

		file, err = os.OpenFile(cbFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err = os.MkdirAll(cbDir, 0755); err != nil {
			Error("Error writing Git Media progress to %s: %s", cbFile, err.Error())
		}

		var prevProgress int
		var progress int

		cb = gitmedia.CopyCallback(func(total int64, written int64) error {
			progress = 0
			if total > 0 {
				progress = int(float64(written) / float64(total) * 100)
			}

			if progress != prevProgress {
				_, err := file.Write([]byte(fmt.Sprintf("smudge %d %s\n", progress, filename)))
				prevProgress = progress
				return err
			}

			return nil
		})
		file.Write([]byte(fmt.Sprintf("smudge 0 %s\n", filename)))
	}

	err = ptr.Smudge(os.Stdout, cb)
	if file != nil {
		file.Close()
	}

	if err != nil {
		ptr.Encode(os.Stdout)
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
