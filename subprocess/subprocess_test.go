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
