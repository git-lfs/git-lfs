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

var localDirSet = NewStringSetFromSlice([]string{".", "./", ".\\"})

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

// FilenamePassesIncludeExcludeFilter returns whether a given filename passes the include / exclude path filters
// Only paths that are in includePaths and outside excludePaths are passed
// If includePaths is empty that filter always passes and the same with excludePaths
// Both path lists support wildcard matches
func FilenamePassesIncludeExcludeFilter(filename string, includePaths, excludePaths []string) bool {
	if len(includePaths) == 0 && len(excludePaths) == 0 {
		return true
	}

	filename = filepath.Clean(filename)
	if len(includePaths) > 0 {
		matched := false
		for _, inc := range includePaths {
			inc = filepath.Clean(inc)

			// Special case local dir, matches all (inc subpaths)
			if _, local := localDirSet[inc]; local {
				matched = true
				break
			}

			matched, _ = filepath.Match(inc, filename)
			if !matched {
				// Also support matching a parent directory without a wildcard
				if strings.HasPrefix(filename, inc+string(filepath.Separator)) {
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
			ex = filepath.Clean(ex)

			// Special case local dir, matches all (inc subpaths)
			if _, local := localDirSet[ex]; local {
				return false
			}

			if matched, _ := filepath.Match(ex, filename); matched {
				return false
			}

			// Also support matching a parent directory without a wildcard
			if strings.HasPrefix(filename, ex+string(filepath.Separator)) {
				return false
			}
		}
	}

	return true
}

// Returned from FastWalk with parent directory context
// This is needed because FastWalk can provide paths out of order so the
// parent dir cannot be implied
type FastWalkInfo struct {
	ParentDir string
	Info      os.FileInfo
}

// FastWalk is a more optimal implementation of filepath.Walk
// It differs in the following ways:
//  * Provides a channel of information instead of using a callback func
//  * Uses goroutines to parallelise large dirs and descent into subdirs
//  * Does not provide sorted output; parents will always be before children but
//    there are no other guarantees. Use parentDir in the FastWalkInfo struct to
//    determine absolute path rather than tracking it yourself like filepath.Walk
//  * Supports include / exclude filters
// Both dir and include/exclude paths can be relative or absolute, but they must
// all be of the same type. includePaths/excludePaths can be nil.
func FastWalk(dir string, includePaths, excludePaths []string) (<-chan FastWalkInfo, <-chan error) {
	return FastWalkWithExcludeFiles(dir, "", includePaths, excludePaths)
}

// FastWalkWithExcludeFiles is like FastWalk but with the additional option to
// load any file named excludeFilename in any directory, and add its contents
// to the excludePaths list for that directory and children.
func FastWalkWithExcludeFiles(dir, excludeFilename string,
	includePaths, excludePaths []string) (<-chan FastWalkInfo, <-chan error) {
	fiChan := make(chan FastWalkInfo, 256)
	errChan := make(chan error, 10)

	dirFi, err := os.Stat(dir)
	if err != nil {
		errChan <- err
		return fiChan, errChan
	}

	go fastWalkItem("", dirFi, excludeFilename, includePaths, excludePaths, fiChan, errChan)

	return fiChan, errChan
}

// FastWalkGitRepo behaves like FastWalkWithExcludeFiles, preconfigured to ignore
// the git repo itself (.git) and to load exclude patterns from .gitignore
func FastWalkGitRepo(dir string) (<-chan FastWalkInfo, <-chan error) {
	excludePaths := []string{".git"}
	return FastWalkWithExcludeFiles(dir, ".gitignore", nil, excludePaths)
}

// Main recursive implementation of fast walk
func fastWalkItem(parentDir string, itemFi os.FileInfo, excludeFilename string,
	includePaths, excludePaths []string, fiChan chan<- FastWalkInfo, errChan chan<- error) {

	fullPath := filepath.Join(parentDir, itemFi.Name())
	if !FilenamePassesIncludeExcludeFilter(fullPath, includePaths, excludePaths) {
		return
	}

	fiChan <- FastWalkInfo{ParentDir: parentDir, Info: itemFi}

	if !itemFi.IsDir() {
		// Nothing more to do if this is not a dir
		return
	}

	if len(excludeFilename) > 0 {
		possibleExcludeFile := filepath.Join(fullPath, excludeFilename)
		if FileExists(possibleExcludeFile) {
			excludePaths = loadExcludeFilename(possibleExcludeFile, excludePaths)
		}
	}

	// The absolute optimal way to scan would be File.Readdirnames but we
	// still need the Stat() to know whether something is a dir, so use
	// File.Readdir instead. Means we can provide os.FileInfo to callers like
	// filepath.Walk as a bonus.
	df, err := os.Open(fullPath)
	if err != nil {
		errChan <- err
		return
	}
	jobSize := 256
	for children, err := df.Readdir(jobSize); err == nil; children, err = df.Readdir(jobSize) {
		// Parallelise all dirs, and chop large dirs into batches of 256
		go func() {
			for _, childFi := range children {
				fastWalkItem(fullPath, childFi, excludeFilename, includePaths, excludePaths, fiChan, errChan)
			}
		}()

	}
	if err != io.EOF {
		errChan <- err
	}
}

// loadExcludeFilename reads the given file in gitignore format and returns a
// revised array of exclude paths if there are any changes.
// If any changes are made a copy of the array is taken so the original is not
// modified
func loadExcludeFilename(filename string, excludePaths []string) []string {
	// TODO
	return excludePaths
}
