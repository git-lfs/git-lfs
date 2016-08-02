// Package tools contains other helper functions too small to justify their own package
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package tools

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// FileOrDirExists determines if a file/dir exists, returns IsDir() results too.
func FileOrDirExists(path string) (exists bool, isDir bool) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, false
	} else {
		return true, fi.IsDir()
	}
}

// FileExists determines if a file (NOT dir) exists.
func FileExists(path string) bool {
	ret, isDir := FileOrDirExists(path)
	return ret && !isDir
}

// DirExists determines if a dir (NOT file) exists.
func DirExists(path string) bool {
	ret, isDir := FileOrDirExists(path)
	return ret && isDir
}

// FileExistsOfSize determines if a file exists and is of a specific size.
func FileExistsOfSize(path string, sz int64) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return false
	}

	return !fi.IsDir() && fi.Size() == sz
}

// ResolveSymlinks ensures that if the path supplied is a symlink, it is
// resolved to the actual concrete path
func ResolveSymlinks(path string) string {
	if len(path) == 0 {
		return path
	}

	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return path
}

// RenameFileCopyPermissions moves srcfile to destfile, replacing destfile if
// necessary and also copying the permissions of destfile if it already exists
func RenameFileCopyPermissions(srcfile, destfile string) error {
	info, err := os.Stat(destfile)
	if os.IsNotExist(err) {
		// no original file
	} else if err != nil {
		return err
	} else {
		if err := os.Chmod(srcfile, info.Mode()); err != nil {
			return fmt.Errorf("can't set filemode on file %q: %v", srcfile, err)
		}
	}

	if err := os.Rename(srcfile, destfile); err != nil {
		return fmt.Errorf("cannot replace %q with %q: %v", destfile, srcfile, err)
	}
	return nil
}

// CleanPaths splits the given `paths` argument by the delimiter argument, and
// then "cleans" that path according to the path.Clean function (see
// https://golang.org/pkg/path#Clean).
// Note always cleans to '/' path separators regardless of platform (git friendly)
func CleanPaths(paths, delim string) (cleaned []string) {
	// If paths is an empty string, splitting it will yield [""], which will
	// become the path ".". To avoid this, bail out if trimmed paths
	// argument is empty.
	if paths = strings.TrimSpace(paths); len(paths) == 0 {
		return
	}

	for _, part := range strings.Split(paths, delim) {
		part = strings.TrimSpace(part)

		cleaned = append(cleaned, path.Clean(part))
	}

	return cleaned
}

// VerifyFileHash reads a file and verifies whether the SHA is correct
// Returns an error if there is a problem
func VerifyFileHash(oid, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := NewLfsContentHash()
	_, err = io.Copy(h, f)
	if err != nil {
		return err
	}

	calcOid := hex.EncodeToString(h.Sum(nil))
	if calcOid != oid {
		return fmt.Errorf("File %q has an invalid hash %s, expected %s", path, calcOid, oid)
	}

	return nil
}
