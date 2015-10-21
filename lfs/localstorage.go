package lfs

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

func ClearTempObjects() {
	filepath.Walk(LocalObjectTempDir, func(path string, info os.FileInfo, err error) error {
		if shouldDeleteTempObject(path, info) {
			return os.RemoveAll(path)
		}

		return err
	})
}

func shouldDeleteTempObject(path string, info os.FileInfo) bool {
	if info.IsDir() {
		return false
	}

	base := filepath.Base(path)
	parts := strings.SplitN(base, "-", 2)
	if len(parts) < 2 {
		tracerx.Printf("Removing invalid tmp object file: %s", path)
		return true
	}

	if FileExists(localMediaPathNoCreate(parts[0])) {
		tracerx.Printf("Removing existing tmp object file: %s", path)
		return true
	}

	if time.Since(info.ModTime()) > (3 * time.Hour) {
		tracerx.Printf("Removing old tmp object file: %s", path)
		return true
	}

	return false
}
