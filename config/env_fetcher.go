package config

import (
	"os"
	"strings"
	"sync"
)

// EnvFetcher is an implementation of the Fetcher type for communicating with
// the system's environment.
//
// It is safe to use across multiple goroutines.
type EnvFetcher struct {
	// vmu guards read/write access to vals
	vmu sync.Mutex
	// vals maintains a local cache of the system's enviornment variables
	// for fast repeat lookups of a given key.
	vals map[string]string
}

// NewEnvFetcher returns a new *EnvFetcher.
func NewEnvFetcher() *EnvFetcher {
	return &EnvFetcher{
		vals: make(map[string]string),
	}
}

// Get returns the value associated with the given key as stored in the local
// cache, or in the operating system's environment variables.
//
// If there was a cache-hit, the value will be returned from the cache, skipping
// a check against os.Getenv. Otherwise, the value will be fetched from the
// system, stored in the cache, and then returned. If no value was present in
// the cache or in the system, an empty string will be returned.
//
// Get is safe to call across multiple goroutines.
func (e *EnvFetcher) Get(key string) (val string) {
	e.vmu.Lock()
	defer e.vmu.Unlock()

	if i, ok := e.vals[key]; ok {
		return i
	}

	v := os.Getenv(key)
	e.vals[key] = v

	return v
}

// Bool returns the boolean state assosicated with a given key, or the value
// "def", if no value was assosicated.
//
// The "boolean state assosicated with a given key" is defined as the
// case-insensitive string comparsion with the following:
//
// 1) true if...
//   "true", "1", "on", "yes", or "t"
// 2) false if...
//   "false", "0", "off", "no", "f", or otherwise.
func (e *EnvFetcher) Bool(key string, def bool) (val bool) {
	s := e.Get(key)
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

// Set replaces a given key-value pair in the cache (if previously present in
// the cache) and in the system's environment variables. It returns an error if
// one was encountered in setting the environment variable, or `nil` if
// successful.
//
// Note: this method is a temporary measure while some of the old tests still
// rely on this mutable behavior.
func (e *EnvFetcher) Set(key, val string) error {
	e.vmu.Lock()
	defer e.vmu.Unlock()

	if _, ok := e.vals[key]; ok {
		e.vals[key] = val
	}

	return os.Setenv(key, val)
}
