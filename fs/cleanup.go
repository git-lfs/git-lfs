package fs

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/git-lfs/git-lfs/tools"
	"github.com/rubyist/tracerx"
)

func (f *Filesystem) cleanupTmp() error {
	tmpdir := f.TempDir()
	if len(tmpdir) == 0 {
		return nil
	}

	// No temporary directory?  No problem.
	if _, err := os.Stat(tmpdir); err != nil && os.IsNotExist(err) {
		return nil
	}

	traversedDirectories := &sync.Map{}

	var walkErr error
	tools.FastWalkDir(tmpdir, func(parentDir string, info os.FileInfo, err error) {
		if err != nil {
			walkErr = err
		}
		if walkErr != nil {
			return
		}
		path := filepath.Join(parentDir, info.Name())
		if info.IsDir() {
			traversedDirectories.Store(path, info)
			return
		}
		parts := strings.SplitN(info.Name(), "-", 2)
		oid := parts[0]
		if len(parts) == 2 && len(oid) == 64 {
			fi, err := os.Stat(f.ObjectPathname(oid))
			if err == nil && !fi.IsDir() {
				tracerx.Printf("Removing existing tmp object file: %s", path)
				os.RemoveAll(path)
				return
			}
		}

		// Don't prune items in a directory younger than an hour.  These
		// items could be hard links to files from other repositories,
		// which would have an older timestamp but which are still in
		// use by some active process.  Exempt the main temporary from
		// this check, since we frequently modify it and we'd never
		// prune otherwise.
		if tmpdir != parentDir {
			var dirInfo os.FileInfo
			entry, ok := traversedDirectories.Load(parentDir)
			if ok {
				dirInfo = entry.(os.FileInfo)
			} else {
				dirInfo, err = os.Stat(parentDir)
				if err != nil {
					return
				}
				traversedDirectories.Store(path, dirInfo)
			}

			if time.Since(dirInfo.ModTime()) <= time.Hour {
				return
			}
		}

		if time.Since(info.ModTime()) > time.Hour {
			tracerx.Printf("Removing old tmp object file: %s", path)
			os.RemoveAll(path)
			return
		}
	})

	return walkErr
}
