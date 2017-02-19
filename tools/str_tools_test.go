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
		"simple":          {"foo bar", []string{"foo", "bar"}},
		"simple trailing": {"foo bar ", []string{"foo", "bar"}},
		"simple leading":  {" foo bar", []string{"foo", "bar"}},

		"single quotes":          {"foo 'bar baz'", []string{"foo", "'bar baz'"}},
		"single quotes trailing": {"foo 'bar baz' ", []string{"foo", "'bar baz'"}},
		"single quotes leading":  {" foo 'bar baz'", []string{"foo", "'bar baz'"}},

		"double quotes":          {"foo \"bar baz\"", []string{"foo", "\"bar baz\""}},
		"double quotes trailing": {"foo \"bar baz\" ", []string{"foo", "\"bar baz\""}},
		"double quotes leading":  {" foo \"bar baz\"", []string{"foo", "\"bar baz\""}},

		"nested single quotes":          {"foo 'bar 'baz''", []string{"foo", "'bar 'baz''"}},
		"nested single quotes trailing": {"foo 'bar 'baz'' ", []string{"foo", "'bar 'baz''"}},
		"nested single quotes leading":  {" foo 'bar 'baz''", []string{"foo", "'bar 'baz''"}},

		"nested double quotes":          {"foo \"bar \"baz\"\"", []string{"foo", "\"bar \"baz\"\""}},
		"nested double quotes trailing": {"foo \"bar \"baz\"\" ", []string{"foo", "\"bar \"baz\"\""}},
		"nested double quotes leading":  {" foo \"bar \"baz\"\"", []string{"foo", "\"bar \"baz\"\""}},

		"mixed quotes":          {"foo 'bar \"baz\"'", []string{"foo", "'bar \"baz\"'"}},
		"mixed quotes trailing": {"foo 'bar \"baz\"' ", []string{"foo", "'bar \"baz\"'"}},
		"mixed quotes leading":  {" foo 'bar \"baz\"'", []string{"foo", "'bar \"baz\"'"}},
	} {
		t.Run(desc, c.Assert)
	}
}
