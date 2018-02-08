package filepathfilter

import (
	"path/filepath"
	"strings"

	"github.com/git-lfs/wildmatch"
)

type Pattern interface {
	Match(filename string) bool
	// String returns a string representation (see: regular expressions) of
	// the underlying pattern used to match filenames against this Pattern.
	String() string
}

type Filter struct {
	include []Pattern
	exclude []Pattern
}

func NewFromPatterns(include, exclude []Pattern) *Filter {
	return &Filter{include: include, exclude: exclude}
}

func New(include, exclude []string) *Filter {
	return NewFromPatterns(
		convertToWildmatch(include),
		convertToWildmatch(exclude))
}

// Include returns the result of calling String() on each Pattern in the
// include set of this *Filter.
func (f *Filter) Include() []string { return wildmatchToString(f.include...) }

// Exclude returns the result of calling String() on each Pattern in the
// exclude set of this *Filter.
func (f *Filter) Exclude() []string { return wildmatchToString(f.exclude...) }

// wildmatchToString maps the given set of Pattern's to a string slice by
// calling String() on each pattern.
func wildmatchToString(ps ...Pattern) []string {
	s := make([]string, 0, len(ps))
	for _, p := range ps {
		s = append(s, p.String())
	}

	return s
}

func (f *Filter) Allows(filename string) bool {
	if f == nil {
		return true
	}

	var matched bool
	for _, inc := range f.include {
		if matched = inc.Match(filename); matched {
			break
		}
	}

	if !matched && len(f.include) > 0 {
		return false
	}

	for _, ex := range f.exclude {
		if ex.Match(filename) {
			return false
		}
	}

	return true
}

type wm struct {
	w    *wildmatch.Wildmatch
	dirs bool
}

func (w *wm) Match(filename string) bool {
	return w.w.Match(w.chomp(filename))
}

func (w *wm) chomp(filename string) string {
	if !w.dirs {
		base := filepath.Base(filename)
		for _, sep := range []string{"/", "\\"} {
			if strings.HasSuffix(filename, sep) {
				base = base + sep
			}
		}
		return base
	}
	return filename
}

func (w *wm) String() string {
	return w.w.String()
}

func NewPattern(p string) Pattern {
	p = filepath.Clean(p)

	// Special case: "*" and "*.*" should match everything to match existing
	// behavior.
	if p == "*" || p == "*.*" {
		p = "**/*"
	}

	w := wildmatch.NewWildmatch(p, wildmatch.SystemCase)
	dirs := strings.Contains(w.String(), string(filepath.Separator))

	return &wm{
		w:    w,
		dirs: dirs,
	}
}

func convertToWildmatch(rawpatterns []string) []Pattern {
	patterns := make([]Pattern, len(rawpatterns))
	for i, raw := range rawpatterns {
		patterns[i] = NewPattern(raw)
	}
	return patterns
}
