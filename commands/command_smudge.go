package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	// smudgeSkip is a command-line flag belonging to the "git-lfs smudge"
	// command specifying whether to skip the smudge process.
	smudgeSkip = false
)

// smudge smudges the given `*lfs.Pointer`, "ptr", and writes its objects
// contents to the `io.Writer`, "to".
//
// If the smudged object did not "pass" the include and exclude filterset, it
// will not be downloaded, and the object will remain a pointer on disk, as if
// the smudge filter had not been applied at all.
//
// Any errors encountered along the way will be returned immediately if they
// were non-fatal, otherwise execution will halt and the process will be
// terminated by using the `commands.Panic()` func.
func smudge(to io.Writer, from io.Reader, filename string, skip bool, filter *filepathfilter.Filter) error {
	ptr, pbuf, perr := lfs.DecodeFrom(from)
	if perr != nil {
		if _, err := io.Copy(to, pbuf); err != nil {
			return errors.Wrap(err, perr.Error())
		}

		return errors.NewNotAPointerError(errors.Errorf(
			"Unable to parse pointer at: %q", filename,
		))
	}

	lfs.LinkOrCopyFromReference(ptr.Oid, ptr.Size)
	cb, file, err := lfs.CopyCallbackFile("smudge", filename, 1, 1)
	if err != nil {
		return err
	}

	download := !skip
	if download {
		download = filter.Allows(filename)
	}

	err = ptr.Smudge(to, filename, download, getTransferManifest(), cb)
	if file != nil {
		file.Close()
	}

	if err != nil {
		ptr.Encode(to)
		// Download declined error is ok to skip if we weren't requesting download
		if !(errors.IsDownloadDeclinedError(err) && !download) {
			LoggedError(err, "Error downloading object: %s (%s)", filename, ptr.Oid)
			if !cfg.SkipDownloadErrors() {
				os.Exit(2)
			}
		}
	}

	return nil
}

func smudgeCommand(cmd *cobra.Command, args []string) {
	requireStdin("This command should be run by the Git 'smudge' filter")
	lfs.InstallHooks(false)

	if !smudgeSkip && cfg.Os.Bool("GIT_LFS_SKIP_SMUDGE", false) {
		smudgeSkip = true
	}
	filter := filepathfilter.New(cfg.FetchIncludePaths(), cfg.FetchExcludePaths())

	if err := smudge(os.Stdout, os.Stdin, smudgeFilename(args), smudgeSkip, filter); err != nil {
		if errors.IsNotAPointerError(err) {
			fmt.Fprintln(os.Stderr, err.Error())
		} else {
			Error(err.Error())
		}
	}
}

func smudgeFilename(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return "<unknown file>"
}

func init() {
	RegisterCommand("smudge", smudgeCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&smudgeSkip, "skip", "s", false, "")
	})
}
