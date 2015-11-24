package lfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

func ClearTempObjects() {
	if len(LocalObjectTempDir) == 0 {
		return
	}

	d, err := os.Open(LocalObjectTempDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening %q to clear old temp files: %s\n", LocalObjectTempDir, err)
		return
	}

	filenames, _ := d.Readdirnames(-1)
	for _, filename := range filenames {
		path := filepath.Join(LocalObjectTempDir, filename)
		if shouldDeleteTempObject(path) {
			os.RemoveAll(path)
		}
	}
}

func shouldDeleteTempObject(path string) bool {
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

	if FileExists(LocalStorage.ObjectPath(oid)) {
		tracerx.Printf("Removing existing tmp object file: %s", path)
		return true
	}

	if time.Since(info.ModTime()) > time.Hour {
		tracerx.Printf("Removing old tmp object file: %s", path)
		return true
	}

	return false
}
