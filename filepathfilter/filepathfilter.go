package filepathfilter

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
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
	return NewFromPatterns(convertToPatterns(include), convertToPatterns(exclude))
}

// Include returns the result of calling String() on each Pattern in the
// include set of this *Filter.
func (f *Filter) Include() []string { return patternsToStrings(f.include...) }

// Exclude returns the result of calling String() on each Pattern in the
// exclude set of this *Filter.
func (f *Filter) Exclude() []string { return patternsToStrings(f.exclude...) }

// patternsToStrings maps the given set of Pattern's to a string slice by
// calling String() on each pattern.
func patternsToStrings(ps ...Pattern) []string {
	s := make([]string, 0, len(ps))
	for _, p := range ps {
		s = append(s, p.String())
	}

	return s
}

func (f *Filter) Allows(filename string) bool {
	_, allowed := f.AllowsPattern(filename)
	return allowed
}

// AllowsPattern returns whether the given filename is permitted by the
// inclusion/exclusion rules of this filter, as well as the pattern that either
// allowed or disallowed that filename.
//
// In special cases, such as a nil `*Filter` receiver, the absence of any
// patterns, or the given filename not being matched by any pattern, the empty
// string "" will be returned in place of the pattern.
func (f *Filter) AllowsPattern(filename string) (pattern string, allowed bool) {
	if f == nil {
		return "", true
	}

	if len(f.include)+len(f.exclude) == 0 {
		return "", true
	}

	cleanedName := filepath.Clean(filename)

	if len(f.include) > 0 {
		matched := false
		for _, inc := range f.include {
			matched = inc.Match(cleanedName)
			if matched {
				pattern = inc.String()
				break
			}
		}
		if !matched {
			return "", false
		}
	}

	if len(f.exclude) > 0 {
		for _, ex := range f.exclude {
			if ex.Match(cleanedName) {
				return ex.String(), false
			}
		}
	}

	return pattern, true
}

func NewPattern(rawpattern string) Pattern {
	cleanpattern := filepath.Clean(rawpattern)

	// Special case local dir, matches all (inc subpaths)
	if _, local := localDirSet[cleanpattern]; local {
		return noOpMatcher{}
	}

	sep := string(filepath.Separator)
	hasPathSep := strings.Contains(cleanpattern, sep)
	ext := filepath.Ext(cleanpattern)
	plen := len(cleanpattern)
	if plen > 1 && !hasPathSep && strings.HasPrefix(cleanpattern, "*") && cleanpattern[1:plen] == ext {
		return &simpleExtPattern{ext: ext}
	}

	// special case * when there are no path separators
	// filepath.Match never allows * to match a path separator, which is correct
	// for gitignore IF the pattern includes a path separator, but not otherwise
	// So *.txt should match in any subdir, as should test*, but sub/*.txt would
	// only match directly in the sub dir
	// Don't need to test cross-platform separators as both cleaned above
	if !hasPathSep && strings.Contains(cleanpattern, "*") {
		pattern := regexp.QuoteMeta(cleanpattern)
		regpattern := fmt.Sprintf("^%s$", strings.Replace(pattern, "\\*", ".*", -1))
		return &pathlessWildcardPattern{
			rawPattern: cleanpattern,
			wildcardRE: regexp.MustCompile(regpattern),
		}
	}

	// Also support ** with path separators
	if hasPathSep && strings.Contains(cleanpattern, "**") {
		pattern := regexp.QuoteMeta(cleanpattern)
		regpattern := fmt.Sprintf("^%s$", strings.Replace(pattern, "\\*\\*", ".*", -1))
		return &doubleWildcardPattern{
			rawPattern: cleanpattern,
			wildcardRE: regexp.MustCompile(regpattern),
		}
	}

	if hasPathSep && strings.HasPrefix(cleanpattern, sep) {
		rel := cleanpattern[1:len(cleanpattern)]
		prefix := rel
		if strings.HasSuffix(rel, sep) {
			rel = rel[0 : len(rel)-1]
		} else {
			prefix += sep
		}

		return &pathPrefixPattern{
			rawPattern: cleanpattern,
			relative:   rel,
			prefix:     prefix,
		}
	}

	return &pathPattern{
		rawPattern: cleanpattern,
		prefix:     cleanpattern + sep,
		suffix:     sep + cleanpattern,
		inner:      sep + cleanpattern + sep,
	}
}

func convertToPatterns(rawpatterns []string) []Pattern {
	patterns := make([]Pattern, len(rawpatterns))
	for i, raw := range rawpatterns {
		patterns[i] = NewPattern(raw)
	}
	return patterns
}

type pathPrefixPattern struct {
	rawPattern string
	relative   string
	prefix     string
}

// Match is a revised version of filepath.Match which makes it behave more
// like gitignore
func (p *pathPrefixPattern) Match(name string) bool {
	if name == p.relative || strings.HasPrefix(name, p.prefix) {
		return true
	}
	matched, _ := filepath.Match(p.rawPattern, name)
	return matched
}

// String returns a string representation of the underlying pattern for which
// this *pathPrefixPattern is matching.
func (p *pathPrefixPattern) String() string {
	return p.rawPattern
}

type pathPattern struct {
	rawPattern string
	prefix     string
	suffix     string
	inner      string
}

// Match is a revised version of filepath.Match which makes it behave more
// like gitignore
func (p *pathPattern) Match(name string) bool {
	if strings.HasPrefix(name, p.prefix) || strings.HasSuffix(name, p.suffix) || strings.Contains(name, p.inner) {
		return true
	}
	matched, _ := filepath.Match(p.rawPattern, name)
	return matched
}

// String returns a string representation of the underlying pattern for which
// this *pathPattern is matching.
func (p *pathPattern) String() string {
	return p.rawPattern
}

type simpleExtPattern struct {
	ext string
}

func (p *simpleExtPattern) Match(name string) bool {
	return strings.HasSuffix(name, p.ext)
}

// String returns a string representation of the underlying pattern for which
// this *simpleExtPattern is matching.
func (p *simpleExtPattern) String() string {
	return fmt.Sprintf("*%s", p.ext)
}

type pathlessWildcardPattern struct {
	rawPattern string
	wildcardRE *regexp.Regexp
}

// Match is a revised version of filepath.Match which makes it behave more
// like gitignore
func (p *pathlessWildcardPattern) Match(name string) bool {
	matched, _ := filepath.Match(p.rawPattern, name)
	// Match the whole of the base name but allow matching in folders if no path
	return matched || p.wildcardRE.MatchString(filepath.Base(name))
}

// String returns a string representation of the underlying pattern for which
// this *pathlessWildcardPattern is matching.
func (p *pathlessWildcardPattern) String() string {
	return p.rawPattern
}

type doubleWildcardPattern struct {
	rawPattern string
	wildcardRE *regexp.Regexp
}

// Match is a revised version of filepath.Match which makes it behave more
// like gitignore
func (p *doubleWildcardPattern) Match(name string) bool {
	matched, _ := filepath.Match(p.rawPattern, name)
	// Match the whole of the base name but allow matching in folders if no path
	return matched || p.wildcardRE.MatchString(name)
}

// String returns a string representation of the underlying pattern for which
// this *doubleWildcardPattern is matching.
func (p *doubleWildcardPattern) String() string {
	return p.rawPattern
}

type noOpMatcher struct {
}

func (n noOpMatcher) Match(name string) bool {
	return true
}

func (n noOpMatcher) String() string {
	return ""
}

var localDirSet = map[string]struct{}{
	".":   struct{}{},
	"./":  struct{}{},
	".\\": struct{}{},
}
