package config

import (
	"os"
	"sync"
)

// OsFetcher is an implementation of the Fetcher type for communicating with
// the system's environment.
//
// It is safe to use across multiple goroutines.
type OsFetcher struct {
	// vmu guards read/write access to vals
	vmu sync.Mutex
	// vals maintains a local cache of the system's environment variables
	// for fast repeat lookups of a given key.
	vals map[string]*string
}

// NewOsFetcher returns a new *OsFetcher.
func NewOsFetcher() *OsFetcher {
	return &OsFetcher{
		vals: make(map[string]*string),
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
func (o *OsFetcher) Get(key string) (val string, ok bool) {
	o.vmu.Lock()
	defer o.vmu.Unlock()

	if i, ok := o.vals[key]; ok {
		if i == nil {
			return "", false
		}
		return *i, true
	}

	v, ok := os.LookupEnv(key)
	if ok {
		o.vals[key] = &v
	} else {
		o.vals[key] = nil
	}

	return v, ok
}

// GetAll implements the `config.Fetcher.GetAll` method by returning, at most, a
// 1-ary set containing the result of `config.OsFetcher.Get()`.
func (o *OsFetcher) GetAll(key string) []string {
	if v, ok := o.Get(key); ok {
		return []string{v}
	}
	return make([]string, 0)
}

func (o *OsFetcher) All() map[string][]string {
	return nil
}
