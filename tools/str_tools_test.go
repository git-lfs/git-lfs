package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type QuotedFieldsTestCase struct {
	Given    string
	Expected []string
}

func (c *QuotedFieldsTestCase) Assert(t *testing.T) {
	actual := QuotedFields(c.Given)

	assert.Equal(t, c.Expected, actual,
		"tools: expected QuotedFields(%q) to equal %#v (was %#v)",
		c.Given, c.Expected, actual,
	)
}

func TestQuotedFields(t *testing.T) {
	for desc, c := range map[string]QuotedFieldsTestCase{
		"simple":          {`foo bar`, []string{"foo", "bar"}},
		"simple trailing": {`foo bar `, []string{"foo", "bar"}},
		"simple leading":  {` foo bar`, []string{"foo", "bar"}},

		"single quotes":          {`foo 'bar baz'`, []string{"foo", "bar baz"}},
		"single quotes trailing": {`foo 'bar baz' `, []string{"foo", "bar baz"}},
		"single quotes leading":  {` foo 'bar baz'`, []string{"foo", "bar baz"}},

		"single quotes empty":          {`foo ''`, []string{"foo", ""}},
		"single quotes trailing empty": {`foo '' `, []string{"foo", ""}},
		"single quotes leading empty":  {` foo ''`, []string{"foo", ""}},

		"double quotes":          {`foo "bar baz"`, []string{"foo", "bar baz"}},
		"double quotes trailing": {`foo "bar baz" `, []string{"foo", "bar baz"}},
		"double quotes leading":  {` foo "bar baz"`, []string{"foo", "bar baz"}},

		"double quotes empty":          {`foo ""`, []string{"foo", ""}},
		"double quotes trailing empty": {`foo "" `, []string{"foo", ""}},
		"double quotes leading empty":  {` foo ""`, []string{"foo", ""}},

		"nested single quotes":          {`foo 'bar 'baz''`, []string{"foo", "bar 'baz'"}},
		"nested single quotes trailing": {`foo 'bar 'baz'' `, []string{"foo", "bar 'baz'"}},
		"nested single quotes leading":  {` foo 'bar 'baz''`, []string{"foo", "bar 'baz'"}},

		"nested single quotes empty":          {`foo 'bar '''`, []string{"foo", "bar ''"}},
		"nested single quotes trailing empty": {`foo 'bar ''' `, []string{"foo", "bar ''"}},
		"nested single quotes leading empty":  {` foo 'bar '''`, []string{"foo", "bar ''"}},

		"nested double quotes":          {`foo "bar "baz""`, []string{"foo", `bar "baz"`}},
		"nested double quotes trailing": {`foo "bar "baz"" `, []string{"foo", `bar "baz"`}},
		"nested double quotes leading":  {` foo "bar "baz""`, []string{"foo", `bar "baz"`}},

		"nested double quotes empty":          {`foo "bar """`, []string{"foo", `bar ""`}},
		"nested double quotes trailing empty": {`foo "bar """ `, []string{"foo", `bar ""`}},
		"nested double quotes leading empty":  {` foo "bar """`, []string{"foo", `bar ""`}},

		"mixed quotes":          {`foo 'bar "baz"'`, []string{"foo", `bar "baz"`}},
		"mixed quotes trailing": {`foo 'bar "baz"' `, []string{"foo", `bar "baz"`}},
		"mixed quotes leading":  {` foo 'bar "baz"'`, []string{"foo", `bar "baz"`}},

		"mixed quotes empty":          {`foo 'bar ""'`, []string{"foo", `bar ""`}},
		"mixed quotes trailing empty": {`foo 'bar ""' `, []string{"foo", `bar ""`}},
		"mixed quotes leading empty":  {` foo 'bar ""'`, []string{"foo", `bar ""`}},
	} {
		t.Log(desc)
		c.Assert(t)
	}
}

func TestLongestReturnsEmptyStringGivenEmptySet(t *testing.T) {
	assert.Equal(t, "", Longest(nil))
}

func TestLongestReturnsLongestString(t *testing.T) {
	assert.Equal(t, "longest", Longest([]string{"short", "longer", "longest"}))
}

func TestLongestReturnsLastStringGivenSameLength(t *testing.T) {
	assert.Equal(t, "baz", Longest([]string{"foo", "bar", "baz"}))
}

func TestRjustRightJustifiesString(t *testing.T) {
	unjust := []string{
		"short",
		"longer",
		"longest",
	}
	expected := []string{
		"  short",
		" longer",
		"longest",
	}

	assert.Equal(t, expected, Rjust(unjust))
}

func TestLjustLeftJustifiesString(t *testing.T) {
	unjust := []string{
		"short",
		"longer",
		"longest",
	}
	expected := []string{
		"short  ",
		"longer ",
		"longest",
	}

	assert.Equal(t, expected, Ljust(unjust))
}

func TestIndentIndentsStrings(t *testing.T) {
	assert.Equal(t, "\tfoo\n\tbar", Indent("foo\nbar"))
}

func TestIndentIndentsSingleLineStrings(t *testing.T) {
	assert.Equal(t, "\tfoo", Indent("foo"))
}

func TestIndentReturnsEmptyStrings(t *testing.T) {
	assert.Equal(t, "", Indent(""))
}

func TestUndentRemovesLeadingWhitespace(t *testing.T) {
	assert.Equal(t, "foo", Undent("\t\t\tfoo"))
	assert.Equal(t, "foo", Undent("foo"))
	assert.Equal(t, "foo", Undent("    foo"))
}

func TestUndentRemovesPreservesLinebreaks(t *testing.T) {
	// No leading space
	assert.Equal(t, "\r\nfoo", Undent("\r\nfoo"))
	assert.Equal(t, "foo\r\n", Undent("foo\r\n"))
	assert.Equal(t, "\r\nfoo\r\n", Undent("\r\nfoo\r\n"))
	assert.Equal(t, "\nfoo", Undent("\nfoo"))
	assert.Equal(t, "foo\n", Undent("foo\n"))
	assert.Equal(t, "\nfoo\n", Undent("\nfoo\n"))

	// Trim leading space
	assert.Equal(t, "\r\nfoo", Undent("\r\n  foo"))
	assert.Equal(t, "foo\r\n", Undent("  foo\r\n"))
	assert.Equal(t, "\r\nfoo\r\n", Undent("\r\n  foo\r\n"))
	assert.Equal(t, "\nfoo", Undent("\n  foo"))
	assert.Equal(t, "foo\n", Undent("  foo\n"))
	assert.Equal(t, "\nfoo\n", Undent("\n  foo\n"))

	// Preserve trailing space
	assert.Equal(t, "\r\nfoo  ", Undent("\r\nfoo  "))
	assert.Equal(t, "foo  \r\n", Undent("foo  \r\n"))
	assert.Equal(t, "\r\nfoo  \r\n", Undent("\r\nfoo  \r\n"))
	assert.Equal(t, "\nfoo  ", Undent("\nfoo  "))
	assert.Equal(t, "foo  \n", Undent("foo  \n"))
	assert.Equal(t, "\nfoo  \n", Undent("\nfoo  \n"))

	// Trim leading space, preserve trailing space
	assert.Equal(t, "\r\nfoo  ", Undent("\r\n  foo  "))
	assert.Equal(t, "foo  \r\n", Undent("  foo  \r\n"))
	assert.Equal(t, "\r\nfoo  \r\n", Undent("\r\n  foo  \r\n"))
	assert.Equal(t, "\nfoo  ", Undent("\n  foo  "))
	assert.Equal(t, "foo  \n", Undent("  foo  \n"))
	assert.Equal(t, "\nfoo  \n", Undent("\n  foo  \n"))
}
