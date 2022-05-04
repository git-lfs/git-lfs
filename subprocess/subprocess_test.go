package subprocess

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type ShellQuoteTestCase struct {
	Given    []string
	Expected []string
}

func (c *ShellQuoteTestCase) Assert(t *testing.T) {
	actual := ShellQuote(c.Given)

	assert.Equal(t, c.Expected, actual,
		"subprocess: expected ShellQuote(%q) to equal %#v (was %#v)",
		c.Given, c.Expected, actual,
	)
}

func TestShellQuote(t *testing.T) {
	for desc, c := range map[string]ShellQuoteTestCase{
		"simple":         {[]string{"foo", "bar", "an_id"}, []string{"foo", "bar", "an_id"}},
		"leading space":  {[]string{" foo", "bar"}, []string{"' foo'", "bar"}},
		"trailing space": {[]string{"foo", "bar "}, []string{"foo", "'bar '"}},
		"internal space": {[]string{"foo bar", "baz quux"}, []string{"'foo bar'", "'baz quux'"}},
		"backslash":      {[]string{`foo\bar`, `b\az`}, []string{`'foo\bar'`, `'b\az'`}},
		"quotes":         {[]string{`foo"bar`, "b'az"}, []string{`'foo"bar'`, "'b'\\''az'"}},
		"mixed quotes":   {[]string{`"foo'ba\"r\"'"`}, []string{`'"foo'\''ba\"r\"'\''"'`}},
	} {
		t.Run(desc, c.Assert)
	}
}

type FormatForShellQuotedArgsTestCase struct {
	GivenCmd     string
	GivenArgs    []string
	ExpectedArgs []string
}

func (c *FormatForShellQuotedArgsTestCase) Assert(t *testing.T) {
	actualCmd, actualArgs := FormatForShellQuotedArgs(c.GivenCmd, c.GivenArgs)

	assert.Equal(t, "sh", actualCmd,
		"subprocess: expected FormatForShell command to equal 'sh' (was #%v)",
		actualCmd)
	assert.Equal(t, c.ExpectedArgs, actualArgs,
		"subprocess: expected FormatForShell(%q, %v) to equal %#v (was %#v)",
		c.GivenCmd, c.GivenArgs, c.ExpectedArgs, actualArgs,
	)
}

func TestFormatForShellQuotedArgs(t *testing.T) {
	for desc, c := range map[string]FormatForShellQuotedArgsTestCase{
		"simple":      {"foo", []string{"bar", "baz"}, []string{"-c", "foo bar baz"}},
		"spaces":      {"foo quux", []string{" bar", "baz "}, []string{"-c", "foo quux ' bar' 'baz '"}},
		"backslashes": {"bin/foo", []string{"\\bar", "b\\az"}, []string{"-c", "bin/foo '\\bar' 'b\\az'"}},
	} {
		t.Run(desc, c.Assert)
	}
}

type FormatForShellTestCase struct {
	GivenCmd     string
	GivenArgs    string
	ExpectedArgs []string
}

func (c *FormatForShellTestCase) Assert(t *testing.T) {
	actualCmd, actualArgs := FormatForShell(c.GivenCmd, c.GivenArgs)

	assert.Equal(t, "sh", actualCmd,
		"subprocess: expected FormatForShell command to equal 'sh' (was #%v)",
		actualCmd)
	assert.Equal(t, c.ExpectedArgs, actualArgs,
		"subprocess: expected FormatForShell(%q, %v) to equal %#v (was %#v)",
		c.GivenCmd, c.GivenArgs, c.ExpectedArgs, actualArgs,
	)
}

func TestFormatForShell(t *testing.T) {
	for desc, c := range map[string]FormatForShellTestCase{
		"simple": {"foo", "bar", []string{"-c", "foo bar"}},
		"spaces": {"foo quux", "bar baz", []string{"-c", "foo quux bar baz"}},
		"quotes": {"bin/foo", "bar \"baz quux\" 'fred wilma'", []string{"-c", "bin/foo bar \"baz quux\" 'fred wilma'"}},
	} {
		t.Run(desc, c.Assert)
	}
}

type FormatPercentSequencesTestCase struct {
	GivenPattern      string
	GivenReplacements map[string]string
	ExpectedString    string
}

func (c *FormatPercentSequencesTestCase) Assert(t *testing.T) {
	actualString := FormatPercentSequences(c.GivenPattern, c.GivenReplacements)

	assert.Equal(t, c.ExpectedString, actualString,
		"subprocess: expected FormatForShell(%q, %v) to equal %q (was %q)",
		c.GivenPattern, c.GivenReplacements, c.ExpectedString, actualString,
	)
}

func TestFormatPercentSequences(t *testing.T) {
	replacements := map[string]string{
		"A": "current",
		"B": "other file",
		"P": "some ' output \" file",
	}
	for desc, c := range map[string]FormatPercentSequencesTestCase{
		"simple":                        {"merge-foo %A", replacements, "merge-foo current"},
		"double-percent":                {"merge-foo %A %%A", replacements, "merge-foo current %A"},
		"spaces":                        {"merge-foo %B", replacements, "merge-foo 'other file'"},
		"weird filename":                {"merge-foo %P", replacements, "merge-foo 'some '\\'' output \" file'"},
		"no patterns":                   {"merge-foo /dev/null", replacements, "merge-foo /dev/null"},
		"pattern adjacent to non-space": {"merge-foo >%B", replacements, "merge-foo >'other file'"},
	} {
		t.Run(desc, c.Assert)
	}
}
