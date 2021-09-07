package gitattr

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/wildmatch/v2"
)

const attrPrefix = "[attr]"

// Line carries a single line from a repository's .gitattributes file, affecting
// a single pattern and applying zero or more attributes.
type Line struct {
	// Pattern is a wildmatch pattern that, when matched, indicates that all
	// of the below attributes (Attrs) should be applied to that tree entry.
	//
	// Pattern is relative to the tree in which the .gitattributes was read
	// from. For example, /.gitattributes affects all blobs in the
	// repository, while /path/to/.gitattributes affects all blobs that are
	// direct or indirect children of /path/to.
	Pattern *wildmatch.Wildmatch
	// Macro is the name of a macro that, when matched, indicates that all
	// of the below attributes (Attrs) should be applied to that tree
	// entry.
	//
	// A given entry will have exactly one of Pattern or Macro set.
	Macro string
	// Attrs is the list of attributes to be applied when the above pattern
	// matches a given filename.
	//
	// It is populated in-order as it was written in the .gitattributes file
	// being read, from left to right.
	Attrs []*Attr
}

// Attr is a single attribute that may be applied to a file.
type Attr struct {
	// K is the name of the attribute. It is commonly, "filter", "diff",
	// "merge", or "text".
	//
	// It will never contain the special "false" shorthand ("-"), or the
	// unspecify declarative ("!").
	K string
	// V is the value held by that attribute. It is commonly "lfs", or
	// "false", indicating the special value given by a "-"-prefixed name.
	V string
	// Unspecified indicates whether or not this attribute was explicitly
	// unset by prefixing the keyname with "!".
	Unspecified bool
}

// ParseLines parses the given io.Reader "r" line-wise as if it were the
// contents of a .gitattributes file.
//
// If an error was encountered, it will be returned and the []*Line should be
// considered unusable.
func ParseLines(r io.Reader) ([]*Line, string, error) {
	var lines []*Line

	splitter := &lineEndingSplitter{}

	scanner := bufio.NewScanner(r)
	scanner.Split(splitter.ScanLines)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if len(text) == 0 {
			continue
		}

		var pattern string
		var applied string
		var macro string

		switch text[0] {
		case '#':
			continue
		case '"':
			var err error
			last := strings.LastIndex(text, "\"")
			if last == 0 {
				return nil, "", errors.Errorf("git/gitattr: unbalanced quote: %s", text)
			}
			pattern, err = strconv.Unquote(text[:last+1])
			if err != nil {
				return nil, "", errors.Wrapf(err, "git/gitattr")
			}
			applied = strings.TrimSpace(text[last+1:])
		default:
			splits := strings.SplitN(text, " ", 2)

			if strings.HasPrefix(splits[0], attrPrefix) {
				macro = splits[0][len(attrPrefix):]
			} else {
				pattern = splits[0]
			}
			if len(splits) == 2 {
				applied = splits[1]
			}
		}

		var attrs []*Attr

		for _, s := range strings.Split(applied, " ") {
			if s == "" {
				continue
			}

			var attr Attr

			if strings.HasPrefix(s, "-") {
				attr.K = strings.TrimPrefix(s, "-")
				attr.V = "false"
			} else if strings.HasPrefix(s, "!") {
				attr.K = strings.TrimPrefix(s, "!")
				attr.Unspecified = true
			} else if eq := strings.Index(s, "="); eq > -1 {
				attr.K = s[:eq]
				attr.V = s[eq+1:]
			} else {
				attr.K = s
				attr.V = "true"
			}

			attrs = append(attrs, &attr)
		}

		var matchPattern *wildmatch.Wildmatch
		if pattern != "" {
			matchPattern = wildmatch.NewWildmatch(pattern,
				wildmatch.Basename, wildmatch.SystemCase,
				wildmatch.GitAttributes,
			)
		}

		lines = append(lines, &Line{
			Macro:   macro,
			Pattern: matchPattern,
			Attrs:   attrs,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, "", err
	}
	return lines, splitter.LineEnding(), nil
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
