package localstorage

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rubyist/tracerx"
)

func (s *LocalStorage) ClearTempObjects() error {
	if len(s.TempDir) == 0 {
		return nil
	}

	d, err := os.Open(s.TempDir)
	if err != nil {
		return err
	}

	filenames, _ := d.Readdirnames(-1)
	for _, filename := range filenames {
		path := filepath.Join(s.TempDir, filename)
		if shouldDeleteTempObject(s, path) {
			os.RemoveAll(path)
		}
	}

	return nil
}

func shouldDeleteTempObject(s *LocalStorage, path string) bool {
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

	fi, err := os.Stat(s.ObjectPath(oid))
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
