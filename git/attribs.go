package git

import (
	"bufio"
	"bytes"
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
	Source *AttributeSource
	// Path also has the 'lockable' attribute
	Lockable bool
}

type AttributeSource struct {
	Path       string
	LineEnding string
}

func (s *AttributeSource) String() string {
	return s.Path
}

// GetAttributePaths returns a list of entries in .gitattributes which are
// configured with the filter=lfs attribute
// workingDir is the root of the working copy
// gitDir is the root of the git repo
func GetAttributePaths(workingDir, gitDir string) []AttributePath {
	paths := make([]AttributePath, 0)

	for _, path := range findAttributeFiles(workingDir, gitDir) {
		attributes, err := os.Open(path)
		if err != nil {
			continue
		}

		relfile, _ := filepath.Rel(workingDir, path)
		reldir := filepath.Dir(relfile)
		source := &AttributeSource{Path: relfile}

		le := &lineEndingSplitter{}
		scanner := bufio.NewScanner(attributes)
		scanner.Split(le.ScanLines)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			if strings.HasPrefix(line, "#") {
				continue
			}

			// Check for filter=lfs (signifying that LFS is tracking
			// this file) or "lockable", which indicates that the
			// file is lockable (and may or may not be tracked by
			// Git LFS).
			if strings.Contains(line, "filter=lfs") ||
				strings.HasSuffix(line, "lockable") {

				fields := strings.Fields(line)
				pattern := fields[0]
				if len(reldir) > 0 {
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
				paths = append(paths, AttributePath{
					Path:     pattern,
					Source:   source,
					Lockable: lockable,
				})
			}
		}

		source.LineEnding = le.LineEnding()
	}

	return paths
}

// copies bufio.ScanLines(), counting LF vs CRLF in a file
type lineEndingSplitter struct {
	LFCount   int
	CRLFCount int
}

func (s *lineEndingSplitter) LineEnding() string {
	if s.CRLFCount > s.LFCount {
		return "\r\n"
	} else if s.LFCount == 0 {
		return ""
	}
	return "\n"
}

func (s *lineEndingSplitter) ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, s.dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

// dropCR drops a terminal \r from the data.
func (s *lineEndingSplitter) dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		s.CRLFCount++
		return data[0 : len(data)-1]
	}
	s.LFCount++
	return data
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
