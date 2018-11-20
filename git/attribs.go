package git

import (
	"os"
	"path/filepath"

	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git/gitattr"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/rubyist/tracerx"
)

const (
	LockableAttrib = "lockable"
	FilterAttrib   = "filter"
)

// AttributePath is a path entry in a gitattributes file which has the LFS filter
type AttributePath struct {
	// Path entry in the attribute file
	Path string
	// The attribute file which was the source of this entry
	Source *AttributeSource
	// Path also has the 'lockable' attribute
	Lockable bool
	// Path is handled by Git LFS (i.e., filter=lfs)
	Tracked bool
}

type AttributeSource struct {
	Path       string
	LineEnding string
}

func (s *AttributeSource) String() string {
	return s.Path
}

// GetRootAttributePaths beahves as GetRootAttributePaths, and loads information
// only from the global gitattributes file.
func GetRootAttributePaths(cfg Env) []AttributePath {
	af, ok := cfg.Get("core.attributesfile")
	if !ok {
		return nil
	}

	// The working directory for the root gitattributes file is blank.
	return attrPaths(af, "")
}

// GetSystemAttributePaths behaves as GetAttributePaths, and loads information
// only from the system gitattributes file, respecting the $PREFIX environment
// variable.
func GetSystemAttributePaths(env Env) []AttributePath {
	prefix, _ := env.Get("PREFIX")
	if len(prefix) == 0 {
		prefix = string(filepath.Separator)
	}

	path := filepath.Join(prefix, "etc", "gitattributes")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	return attrPaths(path, "")
}

// GetAttributePaths returns a list of entries in .gitattributes which are
// configured with the filter=lfs attribute
// workingDir is the root of the working copy
// gitDir is the root of the git repo
func GetAttributePaths(workingDir, gitDir string) []AttributePath {
	paths := make([]AttributePath, 0)

	for _, path := range findAttributeFiles(workingDir, gitDir) {
		paths = append(paths, attrPaths(path, workingDir)...)
	}

	return paths
}

func attrPaths(path, workingDir string) []AttributePath {
	attributes, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer attributes.Close()

	var paths []AttributePath

	relfile, _ := filepath.Rel(workingDir, path)
	reldir := filepath.Dir(relfile)
	source := &AttributeSource{Path: relfile}

	lines, eol, err := gitattr.ParseLines(attributes)
	if err != nil {
		return nil
	}

	for _, line := range lines {
		lockable := false
		tracked := false
		hasFilter := false

		for _, attr := range line.Attrs {
			if attr.K == FilterAttrib {
				hasFilter = true
				tracked = attr.V == "lfs"
			} else if attr.K == LockableAttrib && attr.V == "true" {
				lockable = true
			}
		}

		if !hasFilter && !lockable {
			continue
		}

		pattern := line.Pattern.String()
		if len(reldir) > 0 {
			pattern = filepath.Join(reldir, pattern)
		}

		paths = append(paths, AttributePath{
			Path:     pattern,
			Source:   source,
			Lockable: lockable,
			Tracked:  tracked,
		})
	}

	source.LineEnding = eol

	return paths
}

// GetAttributeFilter returns a list of entries in .gitattributes which are
// configured with the filter=lfs attribute as a file path filter which
// file paths can be matched against
// workingDir is the root of the working copy
// gitDir is the root of the git repo
func GetAttributeFilter(workingDir, gitDir string) *filepathfilter.Filter {
	paths := GetAttributePaths(workingDir, gitDir)
	patterns := make([]filepathfilter.Pattern, 0, len(paths))

	for _, path := range paths {
		// Convert all separators to `/` before creating a pattern to
		// avoid characters being escaped in situations like `subtree\*.md`
		patterns = append(patterns, filepathfilter.NewPattern(filepath.ToSlash(path.Path)))
	}

	return filepathfilter.NewFromPatterns(patterns, nil)
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
