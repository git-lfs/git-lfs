package config

// mapFetcher provides an implementation of the Fetcher interface by wrapping
// the `map[string]string` type.
type mapFetcher map[string]string

func MapFetcher(m map[string]string) Fetcher {
	return mapFetcher(m)
}

// Get implements the func `Fetcher.Get`.
func (m mapFetcher) Get(key string) (val string, ok bool) {
	val, ok = m[key]
	return
}
