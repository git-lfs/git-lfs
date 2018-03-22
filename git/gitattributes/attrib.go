package gitattributes

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/tools"
)

type Attribute struct {
	Type    string
	Value   string
	Negated bool
}

func (a *Attribute) String() string {
	if a.Negated {
		return fmt.Sprintf("!%s", a.Type)
	}
	return fmt.Sprintf("%s=%s", a.Type, a.Value)
}

type Entry struct {
	Pattern    *Pattern
	Attributes []*Attribute
}

type Entries struct {
	Root    string
	Entries []*Entry
}

func (e *Entries) Matching(s string) []*Entry {
	matching := make([]*Entry, 0, len(e.Entries))

	name := strings.TrimPrefix(strings.TrimPrefix(s, strings.TrimPrefix(e.Root, "/")), "/")
	for _, entry := range e.Entries {
		if entry.Pattern.Match(name) {
			matching = append(matching, entry)
		}
	}

	return matching
}

func ParseEntries(root string, r io.Reader) (*Entries, error) {
	var entries []*Entry

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		var line string = strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}

		fields := tools.QuotedFields(line)

		var attrs []*Attribute

		for _, field := range fields[1:] {
			var attr Attribute

			// TODO(@ttaylorr): support commenting with '#' and ';'.

			if strings.HasPrefix(field, "!") || strings.HasPrefix(field, "-") {
				attr.Type = field[1:]
				attr.Negated = true
			} else {
				splits := strings.SplitN(field, "=", 2)
				if len(splits) != 2 {
					return nil, invalid(line)
				}

				attr.Type = splits[0]
				attr.Value = splits[1]
			}

			attrs = append(attrs, &attr)
		}

		var pattern string = fields[0]
		if line[0] != '"' {
			pattern = unescape(pattern)
		}

		entries = append(entries, &Entry{
			Pattern:    NewPattern(pattern),
			Attributes: attrs,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// TODO(@ttaylorr): replace this with git.RootDir()
	repo := "/Users/ttaylorr/Desktop/example"

	return &Entries{
		Entries: entries,
		Root:    strings.TrimPrefix(filepath.Dir(root), repo),
	}, nil
}

func invalid(line string) error {
	return errors.Errorf("git/gitattributes: invalid attr line: %s", line)
}
