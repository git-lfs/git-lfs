package config_test

import (
	"testing"

	. "github.com/git-lfs/git-lfs/config"
	"github.com/stretchr/testify/assert"
)

func TestEnvironmentGetDelegatesToFetcher(t *testing.T) {
	fetcher := MapFetcher(map[string][]string{
		"foo": []string{"bar", "baz"},
	})

	env := EnvironmentOf(fetcher)
	val, ok := env.Get("foo")

	assert.True(t, ok)
	assert.Equal(t, "baz", val)
}

func TestEnvironmentGetAllDelegatesToFetcher(t *testing.T) {
	fetcher := MapFetcher(map[string][]string{
		"foo": []string{"bar", "baz"},
	})

	env := EnvironmentOf(fetcher)
	vals := env.GetAll("foo")

	assert.Equal(t, []string{"bar", "baz"}, vals)
}

func TestEnvironmentUnsetBoolDefault(t *testing.T) {
	env := EnvironmentOf(MapFetcher(nil))
	assert.True(t, env.Bool("unset", true))
}

func TestEnvironmentBoolTruthyConversion(t *testing.T) {
	for _, c := range []EnvironmentConversionTestCase{
		{"", false, GetBoolDefault(false)},

		{"true", true, GetBoolDefault(false)},
		{"1", true, GetBoolDefault(false)},
		{"on", true, GetBoolDefault(false)},
		{"yes", true, GetBoolDefault(false)},
		{"t", true, GetBoolDefault(false)},

		{"false", false, GetBoolDefault(true)},
		{"0", false, GetBoolDefault(true)},
		{"off", false, GetBoolDefault(true)},
		{"no", false, GetBoolDefault(true)},
		{"f", false, GetBoolDefault(true)},
	} {
		c.Assert(t)
	}
}

func TestEnvironmentIntTestCases(t *testing.T) {
	for _, c := range []EnvironmentConversionTestCase{
		{"", 1, GetIntDefault(1)},

		{"1", 1, GetIntDefault(0)},
		{"3", 3, GetIntDefault(0)},

		{"malformed", 7, GetIntDefault(7)},
	} {
		c.Assert(t)
	}
}

type EnvironmentConversionTestCase struct {
	Val      string
	Expected interface{}

	GotFn func(env Environment, key string) interface{}
}

var (
	GetBoolDefault = func(def bool) func(e Environment, key string) interface{} {
		return func(e Environment, key string) interface{} {
			return e.Bool(key, def)
		}
	}

	GetIntDefault = func(def int) func(e Environment, key string) interface{} {
		return func(e Environment, key string) interface{} {
			return e.Int(key, def)
		}
	}
)

func (c *EnvironmentConversionTestCase) Assert(t *testing.T) {
	fetcher := MapFetcher(map[string][]string{
		c.Val: []string{c.Val},
	})

	env := EnvironmentOf(fetcher)
	got := c.GotFn(env, c.Val)

	if c.Expected != got {
		t.Errorf("lfs/config: expected val=%q to be %q (got: %q)", c.Val, c.Expected, got)
	}
}
