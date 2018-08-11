package gitattr

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/wildmatch"
)

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
func ParseLines(r io.Reader) ([]*Line, error) {
	var lines []*Line

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {

		text := strings.TrimSpace(scanner.Text())
		if len(text) == 0 {
			continue
		}

		var pattern string
		var applied string

		switch text[0] {
		case '#':
			continue
		case '"':
			var err error
			last := strings.LastIndex(text, "\"")
			if last == 0 {
				return nil, errors.Errorf("git/gitattr: unbalanced quote: %s", text)
			}
			pattern, err = strconv.Unquote(text[:last+1])
			if err != nil {
				return nil, errors.Wrapf(err, "git/gitattr")
			}
			applied = strings.TrimSpace(text[last+1:])
		default:
			splits := strings.SplitN(text, " ", 2)

			pattern = splits[0]
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
			} else {
				splits := strings.SplitN(s, "=", 2)
				if len(splits) != 2 {
					return nil, errors.Errorf("git/gitattr: malformed attribute: %s", s)
				}
				attr.K = splits[0]
				attr.V = splits[1]
			}

			attrs = append(attrs, &attr)
		}

		lines = append(lines, &Line{
			Pattern: wildmatch.NewWildmatch(pattern,
				wildmatch.Basename, wildmatch.SystemCase,
			),
			Attrs: attrs,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
