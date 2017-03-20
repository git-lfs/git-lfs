package filepathfilter

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type Pattern interface {
	Match(filename string) bool
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

func (f *Filter) Allows(filename string) bool {
	if f == nil {
		return true
	}

	if len(f.include)+len(f.exclude) == 0 {
		return true
	}

	cleanedName := filepath.Clean(filename)

	if len(f.include) > 0 {
		matched := false
		for _, inc := range f.include {
			matched = inc.Match(cleanedName)
			if matched {
				break
			}
		}
		if !matched {
			return false
		}
	}

	if len(f.exclude) > 0 {
		for _, ex := range f.exclude {
			if ex.Match(cleanedName) {
				return false
			}
		}
	}

	return true
}

func NewPattern(rawpattern string) Pattern {
	cleanpattern := filepath.Clean(rawpattern)

	// Special case local dir, matches all (inc subpaths)
	if _, local := localDirSet[cleanpattern]; local {
		return noOpMatcher{}
	}

	hasPathSep := strings.Contains(cleanpattern, string(filepath.Separator))

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
		// Also support ** with path separators
	} else if hasPathSep && strings.Contains(cleanpattern, "**") {
		pattern := regexp.QuoteMeta(cleanpattern)
		regpattern := fmt.Sprintf("^%s$", strings.Replace(pattern, "\\*\\*", ".*", -1))
		return &doubleWildcardPattern{
			rawPattern: cleanpattern,
			wildcardRE: regexp.MustCompile(regpattern),
		}
	} else {
		return &basicPattern{rawPattern: cleanpattern}
	}
}

func convertToPatterns(rawpatterns []string) []Pattern {
	patterns := make([]Pattern, len(rawpatterns))
	for i, raw := range rawpatterns {
		patterns[i] = NewPattern(raw)
	}
	return patterns
}

type basicPattern struct {
	rawPattern string
}

// Match is a revised version of filepath.Match which makes it behave more
// like gitignore
func (p *basicPattern) Match(name string) bool {
	matched, _ := filepath.Match(p.rawPattern, name)
	// Also support matching a parent directory without a wildcard
	return matched || strings.HasPrefix(name, p.rawPattern+string(filepath.Separator))
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

type noOpMatcher struct {
}

func (n noOpMatcher) Match(name string) bool {
	return true
}

var localDirSet = map[string]struct{}{
	".":   struct{}{},
	"./":  struct{}{},
	".\\": struct{}{},
}
