// Package tools contains other helper functions too small to justify their own package
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package tools

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/filepathfilter"
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

// FastWalkCallback is the signature for the callback given to FastWalkGitRepo()
type FastWalkCallback func(parentDir string, info os.FileInfo, err error)

// FastWalkGitRepo is a more optimal implementation of filepath.Walk for a Git
// repo. The callback guaranteed to be called sequentially. The function returns
// once all files and errors have triggered callbacks.
// It differs in the following ways:
//  * Uses goroutines to parallelise large dirs and descent into subdirs
//  * Does not provide sorted output; parents will always be before children but
//    there are no other guarantees. Use parentDir argument in the callback to
//    determine absolute path rather than tracking it yourself
//  * Automatically ignores any .git directories
//  * Respects .gitignore contents and skips ignored files/dirs
func FastWalkGitRepo(dir string, cb FastWalkCallback) {
	// Ignore all git metadata including subrepos
	excludePaths := []filepathfilter.Pattern{
		filepathfilter.NewPattern(".git"),
		filepathfilter.NewPattern(filepath.Join("**", ".git")),
	}

	fileCh := fastWalkWithExcludeFiles(dir, ".gitignore", excludePaths)
	for file := range fileCh {
		cb(file.ParentDir, file.Info, file.Err)
	}
}

// Returned from FastWalk with parent directory context
// This is needed because FastWalk can provide paths out of order so the
// parent dir cannot be implied
type fastWalkInfo struct {
	ParentDir string
	Info      os.FileInfo
	Err       error
}

// fastWalkWithExcludeFiles walks the contents of a dir, respecting
// include/exclude patterns and also loading new exlude patterns from files
// named excludeFilename in directories walked
func fastWalkWithExcludeFiles(dir, excludeFilename string,
	excludePaths []filepathfilter.Pattern) <-chan fastWalkInfo {
	fiChan := make(chan fastWalkInfo, 256)
	go fastWalkFromRoot(dir, excludeFilename, excludePaths, fiChan)
	return fiChan
}

func fastWalkFromRoot(dir string, excludeFilename string,
	excludePaths []filepathfilter.Pattern, fiChan chan<- fastWalkInfo) {

	dirFi, err := os.Stat(dir)
	if err != nil {
		fiChan <- fastWalkInfo{Err: err}
		return
	}

	// This waitgroup will be incremented for each nested goroutine
	var waitg sync.WaitGroup
	fastWalkFileOrDir(filepath.Dir(dir), dirFi, excludeFilename, excludePaths, fiChan, &waitg)
	waitg.Wait()
	close(fiChan)
}

// fastWalkFileOrDir is the main recursive implementation of fast walk
// Sends the file/dir and any contents to the channel so long as it passes the
// include/exclude filter. If a dir, parses any excludeFilename found and updates
// the excludePaths with its content before (parallel) recursing into contents
// Also splits large directories into multiple goroutines.
// Increments waitg.Add(1) for each new goroutine launched internally
func fastWalkFileOrDir(parentDir string, itemFi os.FileInfo, excludeFilename string,
	excludePaths []filepathfilter.Pattern, fiChan chan<- fastWalkInfo, waitg *sync.WaitGroup) {

	fullPath := filepath.Join(parentDir, itemFi.Name())

	if !filepathfilter.NewFromPatterns(nil, excludePaths).Allows(fullPath) {
		return
	}

	fiChan <- fastWalkInfo{ParentDir: parentDir, Info: itemFi}

	if !itemFi.IsDir() {
		// Nothing more to do if this is not a dir
		return
	}

	if len(excludeFilename) > 0 {
		possibleExcludeFile := filepath.Join(fullPath, excludeFilename)
		var err error
		excludePaths, err = loadExcludeFilename(possibleExcludeFile, fullPath, excludePaths)
		if err != nil {
			fiChan <- fastWalkInfo{Err: err}
		}
	}

	// The absolute optimal way to scan would be File.Readdirnames but we
	// still need the Stat() to know whether something is a dir, so use
	// File.Readdir instead. Means we can provide os.FileInfo to callers like
	// filepath.Walk as a bonus.
	df, err := os.Open(fullPath)
	if err != nil {
		fiChan <- fastWalkInfo{Err: err}
		return
	}
	defer df.Close()

	// The number of items in a dir we process in each goroutine
	jobSize := 100
	for children, err := df.Readdir(jobSize); err == nil; children, err = df.Readdir(jobSize) {
		// Parallelise all dirs, and chop large dirs into batches
		waitg.Add(1)
		go func(subitems []os.FileInfo) {
			for _, childFi := range subitems {
				fastWalkFileOrDir(fullPath, childFi, excludeFilename, excludePaths, fiChan, waitg)
			}
			waitg.Done()
		}(children)

	}
	if err != nil && err != io.EOF {
		fiChan <- fastWalkInfo{Err: err}
	}
}

// loadExcludeFilename reads the given file in gitignore format and returns a
// revised array of exclude paths if there are any changes.
// If any changes are made a copy of the array is taken so the original is not
// modified
func loadExcludeFilename(filename, parentDir string, excludePaths []filepathfilter.Pattern) ([]filepathfilter.Pattern, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return excludePaths, nil
		}
		return excludePaths, err
	}
	defer f.Close()

	retPaths := excludePaths
	modified := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip blanks, comments and negations (not supported right now)
		if len(line) == 0 || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}

		if !modified {
			// copy on write
			retPaths = make([]filepathfilter.Pattern, len(excludePaths))
			copy(retPaths, excludePaths)
			modified = true
		}

		path := line
		// Add pattern in context if exclude has separator, or no wildcard
		// Allow for both styles of separator at this point
		if strings.ContainsAny(path, "/\\") ||
			!strings.Contains(path, "*") {
			path = filepath.Join(parentDir, line)
		}
		retPaths = append(retPaths, filepathfilter.NewPattern(path))
	}

	return retPaths, nil
}

// SetFileWriteFlag changes write permissions on a file
// Used to make a file read-only or not. When writeEnabled = false, the write
// bit is removed for all roles. When writeEnabled = true, the behaviour is
// different per platform:
// On Mac & Linux, the write bit is set only on the owner as per default umask.
// All other bits are unaffected.
// On Windows, all the write bits are set since Windows doesn't support Unix permissions.
func SetFileWriteFlag(path string, writeEnabled bool) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	mode := uint32(stat.Mode())

	if (writeEnabled && (mode&0200) > 0) ||
		(!writeEnabled && (mode&0222) == 0) {
		// no change needed
		return nil
	}

	if writeEnabled {
		mode = mode | 0200 // set owner write only
		// Go's own Chmod makes Windows set all though
	} else {
		mode = mode &^ 0222 // disable all write
	}
	return os.Chmod(path, os.FileMode(mode))
}
