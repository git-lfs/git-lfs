package commands

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/git/gitattr"
	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
)

// gitCriticalExcluded checks files that must never become LFS pointers because
// Git reads them directly as text files. This guard is independent of any
// autotracksize setting or user config.
func gitCriticalExcluded(fileName string) bool {
	if fileName == "" {
		return false
	}
	switch filepath.Base(fileName) {
	case ".gitattributes", ".gitignore", ".gitmodules", ".mailmap", ".gitkeep":
		return true
	}
	return false
}

// defaultAutoTrackExclusions are the default patterns for the
// lfs.autotrackexclude config key. These are used when the user has not
// set the key. They are NOT applied if the user provides their own value.
var defaultAutoTrackExclusions = []string{
	".gitlab-ci.yml",
	".github/*",
	"*.md",
	"*.txt",
	"*.cfg",
	"*.ini",
}

// isAutoTrackExcluded checks whether a file should be excluded from automatic
// LFS tracking. It uses the user-configured lfs.autotrackexclude if set,
// otherwise falls back to defaultAutoTrackExclusions.
func isAutoTrackExcluded(fileName string) bool {
	if fileName == "" {
		return false
	}

	base := filepath.Base(fileName)

	patterns := defaultAutoTrackExclusions
	if val, ok := cfg.Git.Get("lfs.autotrackexclude"); ok && val != "" {
		patterns = strings.Fields(val)
	}

	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, fileName); matched {
			return true
		}
	}

	return false
}

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
	// Git-critical files must never become LFS pointers
	if gitCriticalExcluded(fileName) {
		if _, err := tools.Spool(to, from, cfg.TempDir()); err != nil {
			return nil, err
		}
		return nil, nil
	}

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

	autoTrackSize, _ := gitattr.GetAutoTrackSize(cfg.LocalWorkingDir(), fileName)

	if autoTrackSize > 0 {
		tmpfile, err := os.CreateTemp(cfg.TempDir(), "")
		if err != nil {
			return nil, err
		}
		defer func() {
			tmpfile.Close()
			os.Remove(tmpfile.Name())
		}()

		n, err := io.Copy(tmpfile, from)
		if err != nil {
			return nil, err
		}

		if _, err := tmpfile.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}

		// Check if content is already a pointer (already tracked in LFS)
		if ptr, decodeErr := lfs.DecodePointer(tmpfile); decodeErr == nil && ptr != nil {
			if _, err := tmpfile.Seek(0, io.SeekStart); err != nil {
				return nil, err
			}
			if _, err := io.Copy(to, tmpfile); err != nil {
				return nil, err
			}
			return ptr, nil
		}

		// Pass through if excluded or under the autotrack threshold
		if isAutoTrackExcluded(fileName) || n < autoTrackSize {
			if _, err := tmpfile.Seek(0, io.SeekStart); err != nil {
				return nil, err
			}
			if _, err := io.Copy(to, tmpfile); err != nil {
				return nil, err
			}
			return nil, nil
		}

		// File exceeds threshold, re-read from temp file for gf.Clean
		if _, err := tmpfile.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		from = tmpfile
		fileSize = n
	}

	cleaned, err := gf.Clean(from, fileName, fileSize, cb)
	if file != nil {
		file.Close()
	}

	if cleaned != nil {
		defer cleaned.Teardown()
	}

	if errors.IsCleanPointerError(err) {
		_, err = to.Write(errors.GetContext(err, "bytes").([]byte))
		return nil, err
	}

	if err != nil {
		ExitWithError(errors.Wrap(err, tr.Tr.Get("Error cleaning Git LFS object")))
	}

	tmpfile := cleaned.Filename
	mediafile, err := gf.ObjectPath(cleaned.Oid)
	if err != nil {
		Panic(err, tr.Tr.Get("Unable to get local media path."))
	}

	if stat, _ := os.Stat(mediafile); stat != nil {
		if stat.Size() != cleaned.Size && len(cleaned.Pointer.Extensions) == 0 {
			Exit("%s\n%s\n%s", tr.Tr.Get("Files don't match:"), mediafile, tmpfile)
		}
		tracerx.Printf("%s exists", mediafile)
	} else {
		if err := os.Rename(tmpfile, mediafile); err != nil {
			Panic(err, tr.Tr.Get("Unable to move %s to %s", tmpfile, mediafile))
		}

		tracerx.Printf("Writing %s", mediafile)
	}

	_, err = lfs.EncodePointer(to, cleaned.Pointer)
	return cleaned.Pointer, err
}

func cleanCommand(cmd *cobra.Command, args []string) {
	requireStdin(tr.Tr.Get("This command should be run by the Git 'clean' filter"))
	setupRepository()
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
		Error(tr.Tr.Get("Possibly malformed conversion on Windows, see `git lfs help smudge` for more details."))
	}
}

func init() {
	RegisterCommand("clean", cleanCommand, nil)
}
