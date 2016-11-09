package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/github/git-lfs/errors"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/tools"
	"github.com/spf13/cobra"
)

var (
	smudgeInfo = false
	smudgeSkip = false
)

func smudge(to io.Writer, ptr *lfs.Pointer, filename string, skip bool) error {
	cb, file, err := lfs.CopyCallbackFile("smudge", filename, 1, 1)
	if err != nil {
		return err
	}

	download := tools.FilenamePassesIncludeExcludeFilter(filename, cfg.FetchIncludePaths(), cfg.FetchExcludePaths())

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

	// keeps the initial buffer from lfs.DecodePointer
	b := &bytes.Buffer{}
	r := io.TeeReader(os.Stdin, b)

	ptr, perr := lfs.DecodePointer(r)
	if perr != nil {
		mr := io.MultiReader(b, os.Stdin)
		if _, err := io.Copy(os.Stdout, mr); err != nil {
			Panic(err, "Error writing data to stdout:")
		}
		return
	}

	lfs.LinkOrCopyFromReference(ptr.Oid, ptr.Size)

	if smudgeInfo {
		fmt.Fprintln(os.Stderr, "WARNING: 'smudge --info' is deprecated and will be removed in v2.0")
		fmt.Fprintln(os.Stderr, "USE INSTEAD:")
		fmt.Fprintln(os.Stderr, "  $ git lfs pointer --file=path/to/file")
		fmt.Fprintln(os.Stderr, "  $ git lfs ls-files")
		fmt.Fprintln(os.Stderr, "")

		localPath, err := lfs.LocalMediaPath(ptr.Oid)
		if err != nil {
			Exit(err.Error())
		}

		if stat, err := os.Stat(localPath); err != nil {
			Print("%d --", ptr.Size)
		} else {
			Print("%d %s", stat.Size(), localPath)
		}

		return
	}

	if err := smudge(os.Stdout, ptr, smudgeFilename(args, perr), smudgeSkip); err != nil {
		Error(err.Error())
	}
}

func smudgeFilename(args []string, err error) string {
	if len(args) > 0 {
		return args[0]
	}

	if errors.IsSmudgeError(err) {
		return filepath.Base(errors.GetContext(err, "FileName").(string))
	}

	return "<unknown file>"
}

func init() {
	RegisterCommand("smudge", smudgeCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&smudgeInfo, "info", "i", false, "")
		cmd.Flags().BoolVarP(&smudgeSkip, "skip", "s", false, "")
	})
}
