package filepathfilter

import (
	"path/filepath"
	"strings"

	"github.com/git-lfs/wildmatch"
	"github.com/rubyist/tracerx"
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

	var included bool
	for _, inc := range f.include {
		if included = inc.Match(filename); included {
			break
		}
	}

	tracerx.Printf("filepathfilter: rejecting %q via %v", filename, f.include)
	if !included && len(f.include) > 0 {
		return false
	}

	for _, ex := range f.exclude {
		if ex.Match(filename) {
			tracerx.Printf("filepathfilter: rejecting %q via %q", filename, ex.String())
			return false
		}
	}

	tracerx.Printf("filepathfilter: accepting %q", filename)
	return true
}

type wm struct {
	w    *wildmatch.Wildmatch
	p    string
	dirs bool
}

func (w *wm) Match(filename string) bool {
	return w.w.Match(w.chomp(filename))
}

func (w *wm) chomp(filename string) string {
	return strings.TrimSuffix(filename, string(filepath.Separator))
}

func (w *wm) String() string {
	return w.p
}

const (
	sep byte = '/'
)

func NewPattern(p string) Pattern {
	pp := p

	// Special case: the below patterns match anything according to existing
	// behavior.
	switch pp {
	case `*`, `.`, `./`, `.\`:
		pp = join("**", "*")
	}

	dirs := strings.Contains(pp, string(sep))
	rooted := strings.HasPrefix(pp, string(sep))
	wild := strings.Contains(pp, "*")

	if !dirs && !wild {
		// Special case: if pp is a literal string (optionally including
		// a character class), rewrite it is a substring match.
		pp = join("**", pp, "**")
	} else {
		if dirs && !rooted {
			// Special case: if there are any directory separators,
			// rewrite "pp" as a substring match.
			if !wild {
				pp = join("**", pp, "**")
			}
		} else {
			if rooted {
				// Special case: if there are not any directory
				// separators, rewrite "pp" as a substring
				// match.
				pp = join(pp, "**")
			} else {
				// Special case: if there are not any directory
				// separators, rewrite "pp" as a substring
				// match.
				pp = join("**", pp)
			}
		}
	}
	tracerx.Printf("filepathfilter: rewrite %q as %q", p, pp)

	return &wm{
		p: p,
		w: wildmatch.NewWildmatch(
			pp,
			wildmatch.SystemCase,
		),
		dirs: dirs,
	}
}

// join joins path elements together via the separator "sep" and produces valid
// paths without multiple separators (unless multiple separators were included
// in the original paths []string).
func join(paths ...string) string {
	var joined string

	for i, path := range paths {
		joined = joined + path
		if i != len(paths)-1 && !strings.HasSuffix(path, string(sep)) {
			joined = joined + string(sep)
		}
	}

	return joined
}

func convertToWildmatch(rawpatterns []string) []Pattern {
	patterns := make([]Pattern, len(rawpatterns))
	for i, raw := range rawpatterns {
		patterns[i] = NewPattern(raw)
	}
	return patterns
}
