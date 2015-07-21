package commands

import (
	"os"

	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	cleanCmd = &cobra.Command{
		Use:   "clean",
		Short: "Implements the Git clean filter",
		Run:   cleanCommand,
	}
)

func cleanCommand(cmd *cobra.Command, args []string) {
	requireStdin("This command should be run by the Git 'clean' filter")
	lfs.InstallHooks(false)

	var fileName string
	var cb lfs.CopyCallback
	var file *os.File
	var fileSize int64
	if len(args) > 0 {
		fileName = args[0]

		stat, err := os.Stat(fileName)
		if err == nil && stat != nil {
			fileSize = stat.Size()

			localCb, localFile, err := lfs.CopyCallbackFile("clean", fileName, 1, 1)
			if err != nil {
				Error(err.Error())
			} else {
				cb = localCb
				file = localFile
			}
		}
	}

	cleaned, err := lfs.PointerClean(os.Stdin, fileName, fileSize, cb)
	if file != nil {
		file.Close()
	}

	if cleaned != nil {
		defer cleaned.Teardown()
	}

	if cpErr, ok := err.(*lfs.CleanedPointerError); ok {
		os.Stdout.Write(cpErr.Bytes)
		return
	}

	if err != nil {
		Panic(err, "Error cleaning asset.")
	}

	tmpfile := cleaned.Filename
	mediafile, err := lfs.LocalMediaPath(cleaned.Oid)
	if err != nil {
		Panic(err, "Unable to get local media path.")
	}

	if stat, _ := os.Stat(mediafile); stat != nil {
		if stat.Size() != cleaned.Size && len(cleaned.Pointer.Extensions) == 0 {
			Exit("Files don't match:\n%s\n%s", mediafile, tmpfile)
		}
		Debug("%s exists", mediafile)
	} else {
		if err := os.Rename(tmpfile, mediafile); err != nil {
			Panic(err, "Unable to move %s to %s\n", tmpfile, mediafile)
		}

		Debug("Writing %s", mediafile)
	}

	lfs.EncodePointer(os.Stdout, cleaned.Pointer)
}

func init() {
	RootCmd.AddCommand(cleanCmd)
}
