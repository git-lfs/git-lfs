package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
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
func GetAttributePaths() []AttributePath {
	paths := make([]AttributePath, 0)

	for _, path := range findAttributeFiles() {
		attributes, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(attributes)

		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "filter=lfs") {
				fields := strings.Fields(line)
				relfile, _ := filepath.Rel(LocalWorkingDir, path)
				pattern := fields[0]
				if reldir := filepath.Dir(relfile); len(reldir) > 0 {
					pattern = filepath.Join(reldir, pattern)
				}
				lockable := strings.Contains(line, LockableAttrib)
				paths = append(paths, AttributePath{Path: pattern, Source: relfile, Lockable: lockable})
			}
		}
	}

	return paths
}

func findAttributeFiles() []string {
	paths := make([]string, 0)

	repoAttributes := filepath.Join(LocalGitDir, "info", "attributes")
	if info, err := os.Stat(repoAttributes); err == nil && !info.IsDir() {
		paths = append(paths, repoAttributes)
	}

	filepath.Walk(LocalWorkingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (filepath.Base(path) == ".gitattributes") {
			paths = append(paths, path)
		}
		return nil
	})

	// reverse the order of the files so more specific entries are found first
	// when iterating from the front (respects precedence)
	for i, j := 0, len(paths)-1; i < j; i, j = i+1, j-1 {
		paths[i], paths[j] = paths[j], paths[i]
	}

	return paths
}
