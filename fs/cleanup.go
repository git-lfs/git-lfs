package fs

import (
	"os"
	"path/filepath"
	"strings"
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

	var walkErr error
	tools.FastWalkDir(tmpdir, func(parentDir string, info os.FileInfo, err error) {
		if err != nil {
			walkErr = err
		}
		if walkErr != nil || info.IsDir() {
			return
		}
		path := filepath.Join(parentDir, info.Name())
		parts := strings.SplitN(info.Name(), "-", 2)
		oid := parts[0]
		if len(parts) < 2 || len(oid) != 64 {
			tracerx.Printf("Removing invalid tmp object file: %s", path)
			os.RemoveAll(path)
			return
		}

		fi, err := os.Stat(f.ObjectPathname(oid))
		if err == nil && !fi.IsDir() {
			tracerx.Printf("Removing existing tmp object file: %s", path)
			os.RemoveAll(path)
			return
		}

		if time.Since(info.ModTime()) > time.Hour {
			tracerx.Printf("Removing old tmp object file: %s", path)
			os.RemoveAll(path)
			return
		}
	})

	return walkErr
}
