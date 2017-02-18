package tools

import "regexp"

var (
	// quoteFieldRe greedily matches between matching pairs of '', "", or
	// non-word characters.
	quoteFieldRe = regexp.MustCompile("'.+'|\".+\"|\\S+")
)

// QuotedFields is an alternative to strings.Fields (see:
// https://golang.org/pkg/strings#Fields) that respects spaces between matching
// pairs of quotation delimeters.
//
// For instance, the quoted fields of the string "foo bar 'baz etc'" would be:
//   []string{"foo", "bar", "'baz etc'"}
//
// Whereas the same argument given to strings.Fields, would return:
//   []string{"foo", "bar", "'baz", "etc'"}
func QuotedFields(s string) []string {
	return quoteFieldRe.FindAllString(s, -1)
}
