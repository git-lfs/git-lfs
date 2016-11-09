package commands

import (
	"io"
	"os"

	"github.com/github/git-lfs/errors"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/progress"
	"github.com/spf13/cobra"
)

func clean(from io.Reader, to io.Writer, fileName string) error {
	var cb progress.CopyCallback
	var file *os.File
	var fileSize int64

	if len(fileName) > 0 {
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

	cleaned, err := lfs.PointerClean(from, fileName, fileSize, cb)
	if file != nil {
		file.Close()
	}

	if cleaned != nil {
		defer cleaned.Teardown()
	}

	if errors.IsCleanPointerError(err) {
		// If the contents read from the working directory was _already_
		// a pointer, we'll get a `CleanPointerError`, with the context
		// containing the bytes that we should write back out to Git.

		_, err = to.Write(errors.GetContext(err, "bytes").([]byte))
		return err
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

	_, err = lfs.EncodePointer(to, cleaned.Pointer)
	return err
}

func cleanCommand(cmd *cobra.Command, args []string) {
	requireStdin("This command should be run by the Git 'clean' filter")
	lfs.InstallHooks(false)

	var fileName string
	if len(args) > 0 {
		fileName = args[0]
	}

	if err := clean(os.Stdin, os.Stdout, fileName); err != nil {
		Error(err.Error())
	}
}

func init() {
	RegisterCommand("clean", cleanCommand, nil)
}
