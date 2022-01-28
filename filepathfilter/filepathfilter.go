package filepathfilter

import (
	"strings"

	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/git-lfs/wildmatch/v2"
	"github.com/rubyist/tracerx"
)

type Pattern interface {
	Match(filename string) bool
	// String returns a string representation (see: regular expressions) of
	// the underlying pattern used to match filenames against this Pattern.
	String() string
}

type Filter struct {
	include      []Pattern
	exclude      []Pattern
	defaultValue bool
}

type PatternType bool

const (
	GitIgnore     = PatternType(false)
	GitAttributes = PatternType(true)
)

func (p PatternType) String() string {
	if p == GitIgnore {
		return "gitignore"
	}
	return "gitattributes"
}

type options struct {
	defaultValue bool
}

type option func(*options)

// DefaultValue is an option representing the default value of a filepathfilter
// if no patterns match.  If this option is not provided, the default is true.
func DefaultValue(val bool) option {
	return func(args *options) {
		args.defaultValue = val
	}
}

func NewFromPatterns(include, exclude []Pattern, setters ...option) *Filter {
	args := &options{defaultValue: true}
	for _, setter := range setters {
		setter(args)
	}
	return &Filter{include: include, exclude: exclude, defaultValue: args.defaultValue}
}

func New(include, exclude []string, ptype PatternType, setters ...option) *Filter {
	return NewFromPatterns(
		convertToWildmatch(include, ptype),
		convertToWildmatch(exclude, ptype), setters...)
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

	if !included && len(f.include) > 0 {
		tracerx.Printf("filepathfilter: rejecting %q via %v", filename, f.include)
		return false
	}

	// Beyond this point, the only values we can logically return are false
	// or the default value.  If the default is false, then there's no point
	// traversing the exclude patterns because the return value will always
	// be false; we'd do extra work for no functional benefit.
	if !included && !f.defaultValue {
		tracerx.Printf("filepathfilter: rejecting %q", filename)
		return false
	}

	for _, ex := range f.exclude {
		if ex.Match(filename) {
			tracerx.Printf("filepathfilter: rejecting %q via %q", filename, ex.String())
			return false
		}
	}

	// No patterns matched and our default value is true.
	tracerx.Printf("filepathfilter: accepting %q", filename)
	return true
}

type wm struct {
	w *wildmatch.Wildmatch
	p string
}

func (w *wm) Match(filename string) bool {
	return w.w.Match(filename)
}

func (w *wm) String() string {
	return w.p
}

const (
	sep byte = '/'
)

func NewPattern(p string, ptype PatternType) Pattern {
	tracerx.Printf("filepathfilter: creating pattern %q of type %v", p, ptype)

	switch ptype {
	case GitIgnore:
		return &wm{
			p: p,
			w: wildmatch.NewWildmatch(
				p,
				wildmatch.SystemCase,
				wildmatch.Contents,
			),
		}
	case GitAttributes:
		return &wm{
			p: p,
			w: wildmatch.NewWildmatch(
				p,
				wildmatch.SystemCase,
				wildmatch.Basename,
				wildmatch.GitAttributes,
			),
		}
	default:
		panic(tr.Tr.Get("unreachable"))
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

func convertToWildmatch(rawpatterns []string, ptype PatternType) []Pattern {
	patterns := make([]Pattern, len(rawpatterns))
	for i, raw := range rawpatterns {
		patterns[i] = NewPattern(raw, ptype)
	}
	return patterns
}
