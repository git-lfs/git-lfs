package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/git-lfs/tools/humanize"
	"github.com/git-lfs/git-lfs/tq"
	"github.com/spf13/cobra"
)

var (
	// smudgeSkip is a command-line flag belonging to the "git-lfs smudge"
	// command specifying whether to skip the smudge process.
	smudgeSkip = false
)

// delayedSmudge performs a 'delayed' smudge, adding the LFS pointer to the
// `*tq.TransferQueue` "q" if the file is not present locally, passes the given
// filepathfilter, and is not skipped. If the pointer is malformed, or already
// exists, it streams the contents to be written into the working copy to "to".
//
// delayedSmudge returns the number of bytes written, whether the checkout was
// delayed, the *lfs.Pointer that was smudged, and an error, if one occurred.
func delayedSmudge(gf *lfs.GitFilter, s *git.FilterProcessScanner, to io.Writer, from io.Reader, q *tq.TransferQueue, filename string, skip bool, filter *filepathfilter.Filter) (int64, bool, *lfs.Pointer, error) {
	ptr, pbuf, perr := lfs.DecodeFrom(from)
	if perr != nil {
		// Write 'statusFromErr(nil)', even though 'perr != nil', since
		// we are about to write non-delayed smudged contents to "to".
		if err := s.WriteStatus(statusFromErr(nil)); err != nil {
			return 0, false, nil, err
		}

		n, err := tools.Spool(to, pbuf, cfg.TempDir())
		if err != nil {
			return n, false, nil, errors.Wrap(err, perr.Error())
		}

		if n != 0 {
			return 0, false, nil, errors.NewNotAPointerError(errors.Errorf(
				"Unable to parse pointer at: %q", filename,
			))
		}
		return 0, false, nil, nil
	}

	lfs.LinkOrCopyFromReference(cfg, ptr.Oid, ptr.Size)

	path, err := cfg.Filesystem().ObjectPath(ptr.Oid)
	if err != nil {
		return 0, false, nil, err
	}

	if !skip && filter.Allows(filename) {
		if _, statErr := os.Stat(path); statErr != nil {
			q.Add(filename, path, ptr.Oid, ptr.Size)
			return 0, true, ptr, nil
		}

		// Write 'statusFromErr(nil)', since the object is already
		// present in the local cache, we will write the object's
		// contents without delaying.
		if err := s.WriteStatus(statusFromErr(nil)); err != nil {
			return 0, false, nil, err
		}

		n, err := gf.Smudge(to, ptr, filename, false, nil, nil)
		return n, false, ptr, err
	}

	if err := s.WriteStatus(statusFromErr(nil)); err != nil {
		return 0, false, nil, err
	}

	n, err := ptr.Encode(to)
	return int64(n), false, ptr, err
}

// smudge smudges the given `*lfs.Pointer`, "ptr", and writes its objects
// contents to the `io.Writer`, "to".
//
// If the encoded LFS pointer is not parse-able as a pointer, the contents of
// that file will instead be spooled to a temporary location on disk and then
// copied out back to Git. If the pointer file is empty, an empty file will be
// written with no error.
//
// If the smudged object did not "pass" the include and exclude filterset, it
// will not be downloaded, and the object will remain a pointer on disk, as if
// the smudge filter had not been applied at all.
//
// Any errors encountered along the way will be returned immediately if they
// were non-fatal, otherwise execution will halt and the process will be
// terminated by using the `commands.Panic()` func.
func smudge(gf *lfs.GitFilter, to io.Writer, from io.Reader, filename string, skip bool, filter *filepathfilter.Filter) (int64, error) {
	ptr, pbuf, perr := lfs.DecodeFrom(from)
	if perr != nil {
		n, err := tools.Spool(to, pbuf, cfg.TempDir())
		if err != nil {
			return 0, errors.Wrap(err, perr.Error())
		}

		if n != 0 {
			return 0, errors.NewNotAPointerError(errors.Errorf(
				"Unable to parse pointer at: %q", filename,
			))
		}
		return 0, nil
	}

	lfs.LinkOrCopyFromReference(cfg, ptr.Oid, ptr.Size)
	cb, file, err := gf.CopyCallbackFile("download", filename, 1, 1)
	if err != nil {
		return 0, err
	}

	download := !skip
	if download {
		download = filter.Allows(filename)
	}

	n, err := gf.Smudge(to, ptr, filename, download, getTransferManifest(), cb)
	if file != nil {
		file.Close()
	}

	if err != nil {
		ptr.Encode(to)
		// Download declined error is ok to skip if we weren't requesting download
		if !(errors.IsDownloadDeclinedError(err) && !download) {
			var oid string = ptr.Oid
			if len(oid) >= 7 {
				oid = oid[:7]
			}

			LoggedError(err, "Error downloading object: %s (%s): %s", filename, oid, err)
			if !cfg.SkipDownloadErrors() {
				os.Exit(2)
			}
		}
	}

	return n, nil
}

func smudgeCommand(cmd *cobra.Command, args []string) {
	requireStdin("This command should be run by the Git 'smudge' filter")
	installHooks(false)

	if !smudgeSkip && cfg.Os.Bool("GIT_LFS_SKIP_SMUDGE", false) {
		smudgeSkip = true
	}
	filter := filepathfilter.New(cfg.FetchIncludePaths(), cfg.FetchExcludePaths())
	gitfilter := lfs.NewGitFilter(cfg)

	if n, err := smudge(gitfilter, os.Stdout, os.Stdin, smudgeFilename(args), smudgeSkip, filter); err != nil {
		if errors.IsNotAPointerError(err) {
			fmt.Fprintln(os.Stderr, err.Error())
		} else {
			Error(err.Error())
		}
	} else if possiblyMalformedObjectSize(n) {
		fmt.Fprintln(os.Stderr, "Possibly malformed smudge on Windows: see `git lfs help smudge` for more info.")
	}
}

func smudgeFilename(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return "<unknown file>"
}

func possiblyMalformedObjectSize(n int64) bool {
	return n > 4*humanize.Gigabyte
}

func init() {
	RegisterCommand("smudge", smudgeCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&smudgeSkip, "skip", "s", false, "")
	})
}
