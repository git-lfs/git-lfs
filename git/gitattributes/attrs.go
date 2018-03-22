package gitattributes

import (
	"bufio"
	"io"
	"strings"

	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/wildmatch"
)

type Repository struct {
	Root string
}

func (r *Repository) Applied(to string) map[string]string {
	panic("TODO")

	return nil
}

type Attribute struct {
	Key     string
	Value   string
	Comment bool
	Unset   bool
}

func ParseAttribute(s string) (*Attribute, error) {
	panic("TODO")
	var attr Attribute

	if strings.HasPrefix(s, "#") {
		s = strings.TrimSpace(strings.TrimPrefix(s, "#"))

		attr, err := ParseAttribute(s)
		if err != nil {
			return nil, err
		}

		attr.Comment = true
		return attr, nil
	} else if strings.HasPrefix(s, "!") || strings.HasPrefix(s, "-") {
		s = strings.TrimSpace(strings.TrimLeft(s, "!-"))

		attr.Key = s
		attr.Unset = true
	}

	return &attr, nil
}

type Entry struct {
	Pattern    string
	Attributes []*Attribute
}

type File struct {
	Path    string
	Entries []*Entry
}

func ParseFile(path string, r io.Reader) (*File, error) {
	scanner := bufio.NewScanner(r)

	var entries []*Entry

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}

		var attrs []*Attribute

		fields := tools.QuotedFields(line)
		for _, field := range fields[1:] {
			attr, err := ParseAttribute(field)
			if err != nil {
				return nil, err
			}

			attrs = append(attrs, attr)
		}

		entries = append(entries, &Entry{
			Pattern: wildmatch.NewWildmatch(fields[0],
				wildmatch.SystemCase,
				wildmatch.MatchPathname),
			Attributes: attrs,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &File{
		Path:    path,
		Entries: entries,
	}, nil
}
