package config

import (
	"strconv"
	"strings"
)

// An Environment adds additional behavior to a Fetcher, such a type conversion,
// and default values.
//
// `Environment`s are the primary way to communicate with various configuration
// sources, such as the OS environment variables, the `.gitconfig`, and even
// `map[string]string`s.
type Environment interface {
	// Get is shorthand for calling `e.Fetcher.Get(key)`.
	Get(key string) (val string, ok bool)

	// Bool returns the boolean state associated with a given key, or the
	// value "def", if no value was associated.
	//
	// The "boolean state associated with a given key" is defined as the
	// case-insensitive string comparison with the following:
	//
	// 1) true if...
	//   "true", "1", "on", "yes", or "t"
	// 2) false if...
	//   "false", "0", "off", "no", "f", or otherwise.
	Bool(key string, def bool) (val bool)

	// Int returns the int value associated with a given key, or the value
	// "def", if no value was associated.
	//
	// To convert from a the string value attached to a given key,
	// `strconv.Atoi(val)` is called. If `Atoi` returned a non-nil error,
	// then the value "def" will be returned instead.
	//
	// Otherwise, if the value was converted `string -> int` successfully,
	// then it will be returned wholesale.
	Int(key string, def int) (val int)

	// All returns a copy of all the key/value pairs for the current environment.
	All() map[string]string
}

type environment struct {
	// Fetcher is the `environment`'s source of data.
	Fetcher Fetcher
}

// EnvironmentOf creates a new `Environment` initialized with the givne
// `Fetcher`, "f".
func EnvironmentOf(f Fetcher) Environment {
	return &environment{f}
}

func (e *environment) Get(key string) (val string, ok bool) {
	return e.Fetcher.Get(key)
}

func (e *environment) Bool(key string, def bool) (val bool) {
	s, _ := e.Fetcher.Get(key)
	if len(s) == 0 {
		return def
	}

	switch strings.ToLower(s) {
	case "true", "1", "on", "yes", "t":
		return true
	case "false", "0", "off", "no", "f":
		return false
	default:
		return false
	}
}

func (e *environment) Int(key string, def int) (val int) {
	s, _ := e.Fetcher.Get(key)
	if len(s) == 0 {
		return def
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}

	return i
}

func (e *environment) All() map[string]string {
	return e.Fetcher.All()
}
