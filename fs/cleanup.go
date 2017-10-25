package fs

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rubyist/tracerx"
)

func (f *Filesystem) cleanupTmp() error {
	tmpdir := f.TempDir()
	if len(tmpdir) == 0 {
		return nil
	}

	d, err := os.Open(tmpdir)
	if err != nil {
		return err
	}

	filenames, _ := d.Readdirnames(-1)
	for _, filename := range filenames {
		path := filepath.Join(tmpdir, filename)
		if f.shouldDeleteTempObject(path) {
			os.RemoveAll(path)
		}
	}

	return nil
}

func (f *Filesystem) shouldDeleteTempObject(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if info.IsDir() {
		return false
	}

	base := filepath.Base(path)
	parts := strings.SplitN(base, "-", 2)
	oid := parts[0]
	if len(parts) < 2 || len(oid) != 64 {
		tracerx.Printf("Removing invalid tmp object file: %s", path)
		return true
	}

	fi, err := os.Stat(f.ObjectPath(oid))
	if err == nil && !fi.IsDir() {
		tracerx.Printf("Removing existing tmp object file: %s", path)
		return true
	}

	if time.Since(info.ModTime()) > time.Hour {
		tracerx.Printf("Removing old tmp object file: %s", path)
		return true
	}

	return false
}
