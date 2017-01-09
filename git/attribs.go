package git

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/tools"
	"github.com/rubyist/tracerx"
)

const (
	LockableAttrib = "lockable"
)

// AttributePath is a path entry in a gitattributes file which has the LFS filter
type AttributePath struct {
	// Path entry in the attribute file
	Path string
	// The attribute file which was the source of this entry
	Source string
	// Path also has the 'lockable' attribute
	Lockable bool
}

// GetAttributePaths returns a list of entries in .gitattributes which are
// configured with the filter=lfs attribute
// workingDIr is the root of the working copy
// gitDir is the root of the git repo
func GetAttributePaths(workingDir, gitDir string) []AttributePath {
	paths := make([]AttributePath, 0)

	for _, path := range findAttributeFiles(workingDir, gitDir) {
		attributes, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(attributes)

		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "filter=lfs") {
				fields := strings.Fields(line)
				relfile, _ := filepath.Rel(workingDir, path)
				pattern := fields[0]
				if reldir := filepath.Dir(relfile); len(reldir) > 0 {
					pattern = filepath.Join(reldir, pattern)
				}
				// Find lockable flag in any position after pattern to avoid
				// edge case of matching "lockable" to a file pattern
				lockable := false
				for _, f := range fields[1:] {
					if f == LockableAttrib {
						lockable = true
						break
					}
				}
				paths = append(paths, AttributePath{Path: pattern, Source: relfile, Lockable: lockable})
			}
		}
	}

	return paths
}

func findAttributeFiles(workingDir, gitDir string) []string {
	var paths []string

	repoAttributes := filepath.Join(gitDir, "info", "attributes")
	if info, err := os.Stat(repoAttributes); err == nil && !info.IsDir() {
		paths = append(paths, repoAttributes)
	}

	tools.FastWalkGitRepo(workingDir, func(parentDir string, info os.FileInfo, err error) {
		if err != nil {
			tracerx.Printf("Error finding .gitattributes: %v", err)
			return
		}

		if info.IsDir() || info.Name() != ".gitattributes" {
			return
		}
		paths = append(paths, filepath.Join(parentDir, info.Name()))
	})

	// reverse the order of the files so more specific entries are found first
	// when iterating from the front (respects precedence)
	for i, j := 0, len(paths)-1; i < j; i, j = i+1, j-1 {
		paths[i], paths[j] = paths[j], paths[i]
	}

	return paths
}
