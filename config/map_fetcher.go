package config

// mapFetcher provides an implementation of the Fetcher interface by wrapping
// the `map[string]string` type.
type mapFetcher map[string]string

// Get implements the func `Fetcher.Get`.
func (m mapFetcher) Get(key string) (val string) { return m[key] }
