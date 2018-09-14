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
		"tools: expected ShellQuote(%q) to equal %#v (was %#v)",
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
		"tools: expected FormatForShell command to equal 'sh' (was #%v)",
		actualCmd)
	assert.Equal(t, c.ExpectedArgs, actualArgs,
		"tools: expected FormatForShell(%q, %v) to equal %#v (was %#v)",
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
		"tools: expected FormatForShell command to equal 'sh' (was #%v)",
		actualCmd)
	assert.Equal(t, c.ExpectedArgs, actualArgs,
		"tools: expected FormatForShell(%q, %v) to equal %#v (was %#v)",
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
