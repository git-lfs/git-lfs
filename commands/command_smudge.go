package commands

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	smudgeInfo = false
	smudgeSkip = false
	smudgeCmd  = &cobra.Command{
		Use: "smudge",
		Run: smudgeCommand,
	}
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

	cfg := lfs.Config
	download := lfs.FilenamePassesIncludeExcludeFilter(filename, cfg.FetchIncludePaths(), cfg.FetchExcludePaths())

	if smudgeSkip || lfs.Config.GetenvBool("GIT_LFS_SKIP_SMUDGE", false) {
		download = false
	}

	err = ptr.Smudge(os.Stdout, filename, download, cb)
	if file != nil {
		file.Close()
	}

	if err != nil {
		ptr.Encode(os.Stdout)
		// Download declined error is ok to skip if we weren't requesting download
		if !(lfs.IsDownloadDeclinedError(err) && !download) {
			LoggedError(err, "Error accessing media: %s (%s)", filename, ptr.Oid)
			os.Exit(2)
		}
	}
}

func smudgeFilename(args []string, err error) string {
	if len(args) > 0 {
		return args[0]
	}

	if lfs.IsSmudgeError(err) {
		return filepath.Base(lfs.ErrorGetContext(err, "FileName").(string))
	}

	return "<unknown file>"
}

func init() {
	// update man page
	smudgeCmd.Flags().BoolVarP(&smudgeInfo, "info", "i", false, "")
	smudgeCmd.Flags().BoolVarP(&smudgeSkip, "skip", "s", false, "")
	RootCmd.AddCommand(smudgeCmd)
}
