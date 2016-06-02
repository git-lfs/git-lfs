// Package tools contains other helper functions too small to justify their own package
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package tools

import (
	"path/filepath"
	"strings"
)

// CleanPaths splits the given `paths` argument by the delimiter argument, and
// then "cleans" that path according to the filepath.Clean function (see
// https://golang.org/pkg/file/filepath#Clean).
func CleanPaths(paths, delim string) (cleaned []string) {
	// If paths is an empty string, splitting it will yield [""], which will
	// become the filepath ".". To avoid this, bail out if trimmed paths
	// argument is empty.
	if paths = strings.TrimSpace(paths); len(paths) == 0 {
		return
	}

	for _, part := range strings.Split(paths, delim) {
		part = strings.TrimSpace(part)

		cleaned = append(cleaned, filepath.Clean(part))
	}

	return cleaned
}

// CleanPathsDefault cleans the paths contained in the given `paths` argument
// delimited by the `delim`, argument. If an empty set is returned from that
// split, then the fallback argument is returned instead.
func CleanPathsDefault(paths, delim string, fallback []string) []string {
	cleaned := CleanPaths(paths, delim)
	if len(cleaned) == 0 {
		return fallback
	}

	return cleaned
}
