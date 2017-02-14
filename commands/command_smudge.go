package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/tools"
<<<<<<< HEAD
	"github.com/git-lfs/git-lfs/tools/longpathos"
=======
>>>>>>> f8a50160... Merge branch 'master' into no-dwarf-tables
	"github.com/spf13/cobra"
)

var (
	// smudgeInfo is a command-line flag belonging to the "git-lfs smudge"
	// command specifying whether to skip the smudge process and simply
	// print out the info of the files being smudged.
	//
	// As of v1.5.0, it is deprecated.
	smudgeInfo = false
	// smudgeSkip is a command-line flag belonging to the "git-lfs smudge"
	// command specifying whether to skip the smudge process.
	smudgeSkip = false
)

// smudge smudges the given `*lfs.Pointer`, "ptr", and writes its objects
// contents to the `io.Writer`, "to".
//
// If the encoded LFS pointer is not parse-able as a pointer, the contents of
// that file will instead be spooled to a temporary location on disk and then
// copied out back to Git.
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
		if _, err := tools.Spool(to, pbuf); err != nil {
			return errors.Wrap(err, perr.Error())
		}

		return errors.NewNotAPointerError(errors.Errorf(
			"Unable to parse pointer at: %q", filename,
		))
	}

	lfs.LinkOrCopyFromReference(ptr.Oid, ptr.Size)
	if smudgeInfo {
		// only invoked from `filter.lfs.smudge`, not `filter.lfs.process`
		// NOTE: this is deprecated behavior and will be removed in v2.0.0

		fmt.Fprintln(os.Stderr, "WARNING: 'smudge --info' is deprecated and will be removed in v2.0")
		fmt.Fprintln(os.Stderr, "USE INSTEAD:")
		fmt.Fprintln(os.Stderr, "  $ git lfs pointer --file=path/to/file")
		fmt.Fprintln(os.Stderr, "  $ git lfs ls-files")
		fmt.Fprintln(os.Stderr, "")

		localPath, err := lfs.LocalMediaPath(ptr.Oid)
		if err != nil {
			Exit(err.Error())
		}

		if stat, err := longpathos.Stat(localPath); err != nil {
			Print("%d --", ptr.Size)
		} else {
			Print("%d %s", stat.Size(), localPath)
		}

		return nil
	}

	cb, file, err := lfs.CopyCallbackFile("smudge", filename, 1, 1)
	if err != nil {
		return err
	}

	download := filter.Allows(filename)
	if skip || cfg.Os.Bool("GIT_LFS_SKIP_SMUDGE", false) {
		download = false
	}

	err = ptr.Smudge(to, filename, download, TransferManifest(), cb)
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
		cmd.Flags().BoolVarP(&smudgeInfo, "info", "i", false, "")
		cmd.Flags().BoolVarP(&smudgeSkip, "skip", "s", false, "")
	})
}
