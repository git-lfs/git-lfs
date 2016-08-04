package config_test

import (
	"testing"

	. "github.com/github/git-lfs/config"
	"github.com/stretchr/testify/assert"
)

func TestEnvironmentOfReturnsCorrectlyInitializedEnvironment(t *testing.T) {
	fetcher := new(MockFetcher)

	env := EnvironmentOf(fetcher)

	assert.Equal(t, fetcher, env.Fetcher)
}

func TestEnvironmentGetDelegatesToFetcher(t *testing.T) {
	var fetcher MockFetcher
	fetcher.On("Get", "foo").Return("bar").Once()

	env := EnvironmentOf(&fetcher)
	val := env.Get("foo")

	fetcher.AssertExpectations(t)
	assert.Equal(t, "bar", val)
}

func TestEnvironmentBoolTruthyConversion(t *testing.T) {
	for _, c := range []EnvironmentConversionTestCase{
		{"", true, GetBoolDefault(true)},
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

	GotFn func(env *Environment, key string) interface{}
}

var (
	GetBoolDefault = func(def bool) func(e *Environment, key string) interface{} {
		return func(e *Environment, key string) interface{} {
			return e.Bool(key, def)
		}
	}

	GetIntDefault = func(def int) func(e *Environment, key string) interface{} {
		return func(e *Environment, key string) interface{} {
			return e.Int(key, def)
		}
	}
)

func (c *EnvironmentConversionTestCase) Assert(t *testing.T) {
	var fetcher MockFetcher
	fetcher.On("Get", c.Val).Return(c.Val).Once()

	env := EnvironmentOf(&fetcher)
	got := c.GotFn(env, c.Val)

	if c.Expected != got {
		t.Errorf("lfs/config: expected val=%q to be %q (got: %q)", c.Val, c.Expected, got)
	}
	fetcher.AssertExpectations(t)
}
