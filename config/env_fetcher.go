package config

import (
	"os"
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
