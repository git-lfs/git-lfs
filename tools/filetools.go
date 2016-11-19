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
	"regexp"
	"strings"
	"sync"
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

	if len(includePaths) > 0 {
		matched := false
		for _, inc := range includePaths {
			matched = FileMatch(inc, filename)
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
			if FileMatch(ex, filename) {
				return false
			}
		}
	}

	return true
}

// FileMatch is a revised version of filepath.Match which makes it behave more
// like gitignore
func FileMatch(pattern, name string) bool {
	pattern = filepath.Clean(pattern)
	name = filepath.Clean(name)

	// Special case local dir, matches all (inc subpaths)
	if _, local := localDirSet[pattern]; local {
		return true
	}

	if matched, _ := filepath.Match(pattern, name); matched {
		return true
	}

	// special case * when there are no path separators
	// filepath.Match never allows * to match a path separator, which is correct
	// for gitignore IF the pattern includes a path separator, but not otherwise
	// So *.txt should match in any subdir, as should test*, but sub/*.txt would
	// only match directly in the sub dir
	// Don't need to test cross-platform separators as both cleaned above
	if !strings.Contains(pattern, string(filepath.Separator)) &&
		strings.Contains(pattern, "*") {
		pattern = regexp.QuoteMeta(pattern)
		// Match the whole of the base name but allow matching in folders if no path
		basename := filepath.Base(name)
		regpattern := fmt.Sprintf("^%s$", strings.Replace(pattern, "\\*", ".*", -1))
		if regexp.MustCompile(regpattern).MatchString(basename) {
			return true
		}
	}
	// Also support ** with path separators
	if strings.Contains(pattern, string(filepath.Separator)) && strings.Contains(pattern, "**") {
		pattern = regexp.QuoteMeta(pattern)
		regpattern := fmt.Sprintf("^%s$", strings.Replace(pattern, "\\*\\*", ".*", -1))
		if regexp.MustCompile(regpattern).MatchString(name) {
			return true
		}

	}
	// Also support matching a parent directory without a wildcard
	if strings.HasPrefix(name, pattern+string(filepath.Separator)) {
		return true
	}

	return false

}

// Returned from FastWalk with parent directory context
// This is needed because FastWalk can provide paths out of order so the
// parent dir cannot be implied
type FastWalkInfo struct {
	ParentDir string
	Info      os.FileInfo
}

// fastWalkWithExcludeFiles walks the contents of a dir, respecting
// include/exclude patterns and also loading new exlude patterns from files
// named excludeFilename in directories walked
func fastWalkWithExcludeFiles(dir, excludeFilename string,
	includePaths, excludePaths []string) (<-chan FastWalkInfo, <-chan error) {
	fiChan := make(chan FastWalkInfo, 256)
	errChan := make(chan error, 10)

	go fastWalkFromRoot(dir, excludeFilename, includePaths, excludePaths, fiChan, errChan)

	return fiChan, errChan
}

// FastWalkGitRepo is a more optimal implementation of filepath.Walk for a Git repo
// It differs in the following ways:
//  * Provides a channel of information instead of using a callback func
//  * Uses goroutines to parallelise large dirs and descent into subdirs
//  * Does not provide sorted output; parents will always be before children but
//    there are no other guarantees. Use parentDir in the FastWalkInfo struct to
//    determine absolute path rather than tracking it yourself like filepath.Walk
//  * Automatically ignores any .git directories
//  * Respects .gitignore contents and skips ignored files/dirs
func FastWalkGitRepo(dir string) (<-chan FastWalkInfo, <-chan error) {
	// Ignore all git metadata including subrepos
	excludePaths := []string{".git", filepath.Join("**", ".git")}
	return fastWalkWithExcludeFiles(dir, ".gitignore", nil, excludePaths)
}

func fastWalkFromRoot(dir string, excludeFilename string,
	includePaths, excludePaths []string, fiChan chan<- FastWalkInfo, errChan chan<- error) {

	dirFi, err := os.Stat(dir)
	if err != nil {
		errChan <- err
		return
	}

	// This waitgroup will be incremented for each nested goroutine
	var waitg sync.WaitGroup

	fastWalkFileOrDir(filepath.Dir(dir), dirFi, excludeFilename, includePaths, excludePaths, fiChan, errChan, &waitg)

	waitg.Wait()
	close(fiChan)
	close(errChan)

}

// fastWalkFileOrDir is the main recursive implementation of fast walk
// Sends the file/dir and any contents to the channel so long as it passes the
// include/exclude filter. If a dir, parses any excludeFilename found and updates
// the excludePaths with its content before (parallel) recursing into contents
// Also splits large directories into multiple goroutines.
// Increments waitg.Add(1) for each new goroutine launched internally
func fastWalkFileOrDir(parentDir string, itemFi os.FileInfo, excludeFilename string,
	includePaths, excludePaths []string, fiChan chan<- FastWalkInfo, errChan chan<- error,
	waitg *sync.WaitGroup) {

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
			var err error
			excludePaths, err = loadExcludeFilename(possibleExcludeFile, fullPath, excludePaths)
			if err != nil {
				errChan <- err
			}
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
	defer df.Close()
	// The number of items in a dir we process in each goroutine
	jobSize := 100
	for children, err := df.Readdir(jobSize); err == nil; children, err = df.Readdir(jobSize) {
		// Parallelise all dirs, and chop large dirs into batches
		waitg.Add(1)
		go func(subitems []os.FileInfo) {
			for _, childFi := range subitems {
				fastWalkFileOrDir(fullPath, childFi, excludeFilename, includePaths, excludePaths, fiChan, errChan, waitg)
			}
			waitg.Done()
		}(children)

	}
	if err != nil && err != io.EOF {
		errChan <- err
	}

}

// loadExcludeFilename reads the given file in gitignore format and returns a
// revised array of exclude paths if there are any changes.
// If any changes are made a copy of the array is taken so the original is not
// modified
func loadExcludeFilename(filename, parentDir string, excludePaths []string) ([]string, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
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
			retPaths = make([]string, len(excludePaths))
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
		retPaths = append(retPaths, path)
	}

	return retPaths, nil

}
