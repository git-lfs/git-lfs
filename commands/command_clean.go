package commands

import (
	"io"
	"os"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/spf13/cobra"
)

// clean cleans an object read from the given `io.Reader`, "from", and writes
// out a corresponding pointer to the `io.Writer`, "to". If there were any
// errors encountered along the way, they will be returned immediately if the
// error is non-fatal, otherwise they will halt using the built in
// `commands.Panic`.
//
// If fileSize is given as a non-negative (>= 0) integer, that value is used
// with preference to os.Stat(fileName).Size(). If it is given as negative, the
// value from the `stat(1)` call will be used instead.
//
// If the object read from "from" is _already_ a clean pointer, then it will be
// written out verbatim to "to", without trying to make it a pointer again.
func clean(gf *lfs.GitFilter, to io.Writer, from io.Reader, fileName string, fileSize int64) (*lfs.Pointer, error) {
	var cb tools.CopyCallback
	var file *os.File

	if len(fileName) > 0 {
		stat, err := os.Stat(fileName)
		if err == nil && stat != nil {
			if fileSize < 0 {
				fileSize = stat.Size()
			}

			localCb, localFile, err := gf.CopyCallbackFile("clean", fileName, 1, 1)
			if err != nil {
				Error(err.Error())
			} else {
				cb = localCb
				file = localFile
			}
		}
	}

	cleaned, err := gf.Clean(from, fileName, fileSize, cb)
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
		return nil, err
	}

	if err != nil {
		ExitWithError(errors.Wrap(err, "Error cleaning LFS object"))
	}

	tmpfile := cleaned.Filename
	mediafile, err := gf.ObjectPath(cleaned.Oid)
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
	return cleaned.Pointer, err
}

func cleanCommand(cmd *cobra.Command, args []string) {
	requireStdin("This command should be run by the Git 'clean' filter")
	installHooks(false)

	var fileName string
	if len(args) > 0 {
		fileName = args[0]
	}

	gitfilter := lfs.NewGitFilter(cfg)
	ptr, err := clean(gitfilter, os.Stdout, os.Stdin, fileName, -1)
	if err != nil {
		Error(err.Error())
	}

	if ptr != nil && possiblyMalformedObjectSize(ptr.Size) {
		Error("Possibly malformed conversion on Windows, see `git lfs help smudge` for more details.")
	}
}

func init() {
	RegisterCommand("clean", cleanCommand, nil)
}
