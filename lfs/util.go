package lfs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/progress"
	"github.com/git-lfs/git-lfs/tools"
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

type PathConverter interface {
	Convert(string) string
}

// Convert filenames expressed relative to the root of the repo relative to the
// current working dir. Useful when needing to calling git with results from a rooted command,
// but the user is in a subdir of their repo
// Pass in a channel which you will fill with relative files & receive a channel which will get results
func NewRepoToCurrentPathConverter() (PathConverter, error) {
	r, c, p, err := pathConverterArgs()
	if err != nil {
		return nil, err
	}

	return &repoToCurrentPathConverter{
		repoDir:     r,
		currDir:     c,
		passthrough: p,
	}, nil
}

type repoToCurrentPathConverter struct {
	repoDir     string
	currDir     string
	passthrough bool
}

func (p *repoToCurrentPathConverter) Convert(filename string) string {
	if p.passthrough {
		return filename
	}

	abs := filepath.Join(p.repoDir, filename)
	rel, err := filepath.Rel(p.currDir, abs)
	if err != nil {
		// Use absolute file instead
		return abs
	} else {
		return rel
	}
}

// Convert filenames expressed relative to the current directory to be
// relative to the repo root. Useful when calling git with arguments that requires them
// to be rooted but the user is in a subdir of their repo & expects to use relative args
// Pass in a channel which you will fill with relative files & receive a channel which will get results
func NewCurrentToRepoPathConverter() (PathConverter, error) {
	r, c, p, err := pathConverterArgs()
	if err != nil {
		return nil, err
	}

	return &currentToRepoPathConverter{
		repoDir:     r,
		currDir:     c,
		passthrough: p,
	}, nil
}

type currentToRepoPathConverter struct {
	repoDir     string
	currDir     string
	passthrough bool
}

func (p *currentToRepoPathConverter) Convert(filename string) string {
	if p.passthrough {
		return filename
	}

	var abs string
	if filepath.IsAbs(filename) {
		abs = tools.ResolveSymlinks(filename)
	} else {
		abs = filepath.Join(p.currDir, filename)
	}
	reltoroot, err := filepath.Rel(p.repoDir, abs)
	if err != nil {
		// Can't do this, use absolute as best fallback
		return abs
	} else {
		return reltoroot
	}
}

func pathConverterArgs() (string, string, bool, error) {
	currDir, err := os.Getwd()
	if err != nil {
		return "", "", false, fmt.Errorf("Unable to get working dir: %v", err)
	}
	currDir = tools.ResolveSymlinks(currDir)
	return config.LocalWorkingDir, currDir, config.LocalWorkingDir == currDir, nil
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
