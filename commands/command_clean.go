package commands

import (
	"os"

	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
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

	var filename string
	var cb lfs.CopyCallback
	var file *os.File
	var fileSize int64
	if len(args) > 0 {
		filename = args[0]

		stat, err := os.Stat(filename)
		if err == nil && stat != nil {
			fileSize = stat.Size()

			localCb, localFile, err := lfs.CopyCallbackFile("clean", filename, 1, 1)
			if err != nil {
				Error(err.Error())
			} else {
				cb = localCb
				file = localFile
			}
		}
	}

	cleaned, err := lfs.PointerClean(os.Stdin, fileSize, cb)
	if file != nil {
		file.Close()
	}

	if cleaned != nil {
		cleaned.Close()
		defer cleaned.Teardown()
	}

	if cpErr, ok := err.(*lfs.CleanedPointerError); ok {
		os.Stdout.Write(cpErr.Bytes)
		return
	}

	if err != nil {
		Panic(err, "Error cleaning asset.")
	}

	tmpfile := cleaned.File.Name()
	mediafile, err := lfs.LocalMediaPath(cleaned.Oid)
	if err != nil {
		Panic(err, "Unable to get local media path.")
	}

	if stat, _ := os.Stat(mediafile); stat != nil {
		if stat.Size() != cleaned.Size {
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
