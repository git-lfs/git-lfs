package gitattributes

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/wildmatch"
	"github.com/rubyist/tracerx"
)

type Repository struct {
	Root string
}

func NewRepository(root string) *Repository {
	return &Repository{
		Root: root,
	}
}

func (r *Repository) Applied(to string) map[string]string {
	applied := make(map[string]string)
	dirs := strings.Split(filepath.Dir(filepath.ToSlash(to)), "/")

	for i := 0; i <= len(dirs); i++ {
		parent := strings.Join(dirs[:i], "/")
		fname := strings.Join([]string{
			r.Root, parent, ".gitattributes",
		}, "/")

		f, err := os.Open(fname)
		if err != nil {
			if !os.IsNotExist(err) {
				tracerx.Printf("git/gitattributes: could not open %s: %v",
					fname, err)
			}
			continue
		}

		root := filepath.Dir(r.relativize(f.Name()))

		file, err := ParseFile(root, f)
		if err != nil {
			tracerx.Printf("git/gitattributes: could not parse file: %s: %v",
				root, err)
		}

		f.Close()

		entries := file.Applied(to)

		for _, entry := range entries {
			if attr.Unset {
				delete(applied, entry.Key)
			} else if !attr.Comment {
				applied[attr.Key] = attr.Value
			}
		}
	}

	return nil
}

func (r *Repository) relativize(name string) string {
	return strings.Trim(name, r.Root)
}

type Attribute struct {
	Key     string
	Value   string
	Comment bool
	Unset   bool
}

func ParseAttribute(s string) (*Attribute, error) {
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

		attr, err := ParseAttribute(S)
		attr.Unset = true

		return attr, nil
	} else {
		splits := strings.SplitN(s, "=", 2)

		attr := &Attribute{
			Key: splits[0],
		}
		if len(splits) > 1 {
			attr.Value = splits[1]
		}

		return attr, nil
	}
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

func (f *File) Applied(to string) []*Entry {
	p = strings.TrimPrefix(strings.TrimPrefix(to, f.Path), "/")

	applied := make([]*Entry, 0, len(f.Entries))
	for _, entry := range f.Entries {
		if entry.Pattern.Match(p) {
			applied = append(applied, entry)
		}
	}

	return applied
}
