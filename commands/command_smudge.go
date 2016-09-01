package commands

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/github/git-lfs/errors"
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	smudgeInfo = false
	smudgeSkip = false
)

func smudgeCommand(cmd *cobra.Command, args []string) {
	requireStdin("This command should be run by the Git 'smudge' filter")
	lfs.InstallHooks(false)

	// keeps the initial buffer from lfs.DecodePointer
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

	lfs.LinkOrCopyFromReference(ptr.Oid, ptr.Size)

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

	download := lfs.FilenamePassesIncludeExcludeFilter(filename, cfg.FetchIncludePaths(), cfg.FetchExcludePaths())

	if smudgeSkip || cfg.Os.Bool("GIT_LFS_SKIP_SMUDGE", false) {
		download = false
	}

	err = ptr.Smudge(os.Stdout, filename, download, TransferManifest(), cb)
	if file != nil {
		file.Close()
	}

	if err != nil {
		ptr.Encode(os.Stdout)
		// Download declined error is ok to skip if we weren't requesting download
		if !(errors.IsDownloadDeclinedError(err) && !download) {
			LoggedError(err, "Error downloading object: %s (%s)", filename, ptr.Oid)
			if !cfg.SkipDownloadErrors() {
				os.Exit(2)
			}
		}
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
