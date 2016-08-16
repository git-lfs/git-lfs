package lfs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/progress"
	"github.com/github/git-lfs/tools"
)

type Platform int

const (
	PlatformWindows      = Platform(iota)
	PlatformLinux        = Platform(iota)
	PlatformOSX          = Platform(iota)
	PlatformOther        = Platform(iota) // most likely a *nix variant e.g. freebsd
	PlatformUndetermined = Platform(iota)
)

var currentPlatform = PlatformUndetermined

func CopyCallbackFile(event, filename string, index, totalFiles int) (progress.CopyCallback, *os.File, error) {
	logPath, _ := config.Config.Os.Get("GIT_LFS_PROGRESS")
	if len(logPath) == 0 || len(filename) == 0 || len(event) == 0 {
		return nil, nil, nil
	}

	if !filepath.IsAbs(logPath) {
		return nil, nil, fmt.Errorf("GIT_LFS_PROGRESS must be an absolute path")
	}

	cbDir := filepath.Dir(logPath)
	if err := os.MkdirAll(cbDir, 0755); err != nil {
		return nil, nil, wrapProgressError(err, event, logPath)
	}

	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, file, wrapProgressError(err, event, logPath)
	}

	var prevWritten int64

	cb := progress.CopyCallback(func(total int64, written int64, current int) error {
		if written != prevWritten {
			_, err := file.Write([]byte(fmt.Sprintf("%s %d/%d %d/%d %s\n", event, index, totalFiles, written, total, filename)))
			file.Sync()
			prevWritten = written
			return wrapProgressError(err, event, logPath)
		}

		return nil
	})

	return cb, file, nil
}

func wrapProgressError(err error, event, filename string) error {
	if err != nil {
		return fmt.Errorf("Error writing Git LFS %s progress to %s: %s", event, filename, err.Error())
	}

	return nil
}

var localDirSet = tools.NewStringSetFromSlice([]string{".", "./", ".\\"})

// Return whether a given filename passes the include / exclude path filters
// Only paths that are in includePaths and outside excludePaths are passed
// If includePaths is empty that filter always passes and the same with excludePaths
// Both path lists support wildcard matches
func FilenamePassesIncludeExcludeFilter(filename string, includePaths, excludePaths []string) bool {
	if len(includePaths) == 0 && len(excludePaths) == 0 {
		return true
	}

	// For Win32, because git reports files with / separators
	cleanfilename := filepath.Clean(filename)
	if len(includePaths) > 0 {
		matched := false
		for _, inc := range includePaths {
			// Special case local dir, matches all (inc subpaths)
			if _, local := localDirSet[inc]; local {
				matched = true
				break
			}
			matched, _ = filepath.Match(inc, filename)
			if !matched && IsWindows() {
				// Also Win32 match
				matched, _ = filepath.Match(inc, cleanfilename)
			}
			if !matched {
				// Also support matching a parent directory without a wildcard
				if strings.HasPrefix(cleanfilename, inc+string(filepath.Separator)) {
					matched = true
				}
			}
			if matched {
				break
			}

		}
		if !matched {
			return false
		}
	}

	if len(excludePaths) > 0 {
		for _, ex := range excludePaths {
			// Special case local dir, matches all (inc subpaths)
			if _, local := localDirSet[ex]; local {
				return false
			}
			matched, _ := filepath.Match(ex, filename)
			if !matched && IsWindows() {
				// Also Win32 match
				matched, _ = filepath.Match(ex, cleanfilename)
			}
			if matched {
				return false
			}
			// Also support matching a parent directory without a wildcard
			if strings.HasPrefix(cleanfilename, ex+string(filepath.Separator)) {
				return false
			}

		}
	}

	return true
}

func GetPlatform() Platform {
	if currentPlatform == PlatformUndetermined {
		switch runtime.GOOS {
		case "windows":
			currentPlatform = PlatformWindows
		case "linux":
			currentPlatform = PlatformLinux
		case "darwin":
			currentPlatform = PlatformOSX
		default:
			currentPlatform = PlatformOther
		}
	}
	return currentPlatform
}

// Convert filenames expressed relative to the root of the repo relative to the
// current working dir. Useful when needing to calling git with results from a rooted command,
// but the user is in a subdir of their repo
// Pass in a channel which you will fill with relative files & receive a channel which will get results
func ConvertRepoFilesRelativeToCwd(repochan <-chan string) (<-chan string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Unable to get working dir: %v", err)
	}
	wd = tools.ResolveSymlinks(wd)

	// Early-out if working dir is root dir, same result
	passthrough := false
	if config.LocalWorkingDir == wd {
		passthrough = true
	}

	outchan := make(chan string, 1)

	go func() {
		for f := range repochan {
			if passthrough {
				outchan <- f
				continue
			}
			abs := filepath.Join(config.LocalWorkingDir, f)
			rel, err := filepath.Rel(wd, abs)
			if err != nil {
				// Use absolute file instead
				outchan <- abs
			} else {
				outchan <- rel
			}
		}
		close(outchan)
	}()

	return outchan, nil
}

// Convert filenames expressed relative to the current directory to be
// relative to the repo root. Useful when calling git with arguments that requires them
// to be rooted but the user is in a subdir of their repo & expects to use relative args
// Pass in a channel which you will fill with relative files & receive a channel which will get results
func ConvertCwdFilesRelativeToRepo(cwdchan <-chan string) (<-chan string, error) {
	curdir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Could not retrieve current directory: %v", err)
	}
	// Make sure to resolve symlinks
	curdir = tools.ResolveSymlinks(curdir)

	// Early-out if working dir is root dir, same result
	passthrough := false
	if config.LocalWorkingDir == curdir {
		passthrough = true
	}

	outchan := make(chan string, 1)
	go func() {
		for p := range cwdchan {
			if passthrough {
				outchan <- p
				continue
			}
			var abs string
			if filepath.IsAbs(p) {
				abs = tools.ResolveSymlinks(p)
			} else {
				abs = filepath.Join(curdir, p)
			}
			reltoroot, err := filepath.Rel(config.LocalWorkingDir, abs)
			if err != nil {
				// Can't do this, use absolute as best fallback
				outchan <- abs
			} else {
				outchan <- reltoroot
			}
		}
		close(outchan)
	}()

	return outchan, nil

}

// Are we running on Windows? Need to handle some extra path shenanigans
func IsWindows() bool {
	return GetPlatform() == PlatformWindows
}

func CopyFileContents(src string, dst string) error {
	tmp, err := ioutil.TempFile(TempDir(), filepath.Base(dst))
	if err != nil {
		return err
	}
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	_, err = io.Copy(tmp, in)
	if err != nil {
		return err
	}
	err = tmp.Close()
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), dst)
}

func LinkOrCopy(src string, dst string) error {
	if src == dst {
		return nil
	}
	err := os.Link(src, dst)
	if err == nil {
		return err
	}
	return CopyFileContents(src, dst)
}
