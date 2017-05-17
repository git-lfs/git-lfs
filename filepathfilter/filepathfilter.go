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
		regpattern := fmt.Sprintf(`\A%s\z`, strings.Replace(pattern, `\*`, ".*", -1))
		return &baseMatchingPattern{
			rawPattern: cleanpattern,
			regex:      regexp.MustCompile(regpattern),
		}
	}

	var regpattern string

	// Also support ** with path separators
	if hasPathSep && strings.Contains(cleanpattern, "**") {
		pattern := regexp.QuoteMeta(cleanpattern)
		regpattern = fmt.Sprintf(`\A%s\z`, strings.Replace(pattern, `\*\*`, ".*", -1))
	} else {
		regpattern = fmt.Sprintf(`(\A|%s)%s(%s|\z)`,
			string(filepath.Separator),
			cleanpattern,
			string(filepath.Separator),
		)
	}

	return &nameMatchingPattern{
		rawPattern: cleanpattern,
		regex:      regexp.MustCompile(regpattern),
	}
}

func convertToPatterns(rawpatterns []string) []Pattern {
	patterns := make([]Pattern, len(rawpatterns))
	for i, raw := range rawpatterns {
		patterns[i] = NewPattern(raw)
	}
	return patterns
}

type nameMatchingPattern struct {
	rawPattern string
	regex      *regexp.Regexp
}

// Match is a revised version of filepath.Match which makes it behave more
// like gitignore
func (p *nameMatchingPattern) Match(name string) bool {
	matched, _ := filepath.Match(p.rawPattern, name)
	// Match the whole of the base name but allow matching in folders if no path
	return matched || p.regex.MatchString(name)
}

type baseMatchingPattern struct {
	rawPattern string
	regex      *regexp.Regexp
}

// Match is a revised version of filepath.Match which makes it behave more
// like gitignore
func (p *baseMatchingPattern) Match(name string) bool {
	matched, _ := filepath.Match(p.rawPattern, name)
	// Match the whole of the base name but allow matching in folders if no path
	return matched || p.regex.MatchString(filepath.Base(name))
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
