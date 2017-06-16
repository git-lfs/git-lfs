package tools

import (
	"regexp"
	"strings"
)

var (
	// quoteFieldRe greedily matches between matching pairs of '', "", or
	// non-word characters.
	quoteFieldRe = regexp.MustCompile("'(.*)'|\"(.*)\"|(\\S*)")
)

// QuotedFields is an alternative to strings.Fields (see:
// https://golang.org/pkg/strings#Fields) that respects spaces between matching
// pairs of quotation delimeters.
//
// For instance, the quoted fields of the string "foo bar 'baz etc'" would be:
//   []string{"foo", "bar", "baz etc"}
//
// Whereas the same argument given to strings.Fields, would return:
//   []string{"foo", "bar", "'baz", "etc'"}
func QuotedFields(s string) []string {
	submatches := quoteFieldRe.FindAllStringSubmatch(s, -1)
	out := make([]string, 0, len(submatches))

	for _, matches := range submatches {
		// if a leading or trailing space is found, ignore that
		if matches[0] == "" {
			continue
		}

		// otherwise, find the first non-empty match (inside balanced
		// quotes, or a space-delimited string)
		var str string
		for _, m := range matches[1:] {
			if len(m) > 0 {
				str = m
				break
			}
		}

		out = append(out, str)
	}

	return out
}

// Ljust returns a copied string slice where each element is left justified to
// match the width of the longest element in the set.
func Ljust(strs []string) []string {
	llen := len(Longest(strs))

	dup := make([]string, len(strs), cap(strs))
	copy(dup, strs)

	for i, str := range strs {
		width := MaxInt(0, llen-len(str))
		padding := strings.Repeat(" ", width)

		dup[i] = str + padding
	}

	return dup
}

// Rjust returns a copied string slice where each element is right justified to
// match the width of the longest element in the set.
func Rjust(strs []string) []string {
	llen := len(Longest(strs))

	dup := make([]string, len(strs), cap(strs))
	copy(dup, strs)

	for i, str := range strs {
		width := MaxInt(0, llen-len(str))
		padding := strings.Repeat(" ", width)

		dup[i] = padding + str
	}

	return dup
}

// Longest returns the longest element in the string slice in O(n) time and O(1)
// space. If strs is empty or nil, an empty string will be returned.
func Longest(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	var longest string
	var llen int
	for _, str := range strs {
		if len(str) >= llen {
			longest = str
			llen = len(longest)
		}
	}

	return longest
}
