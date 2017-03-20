package tools

import "regexp"

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
