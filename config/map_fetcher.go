package config

import "strconv"

// mapFetcher provides an implementation of the Fetcher interface by wrapping
// the `map[string]string` type.
type mapFetcher map[string]string

// Get implements the func `Fetcher.Get`.
func (m mapFetcher) Get(key string) (val string) { return m[key] }

// Bool implements the function Fetcher.Bool.
//
// NOTE: It exists as a temporary measure before the Environment type is
// abstracted (see github/git-lfs#1415).
func (m mapFetcher) Bool(key string, def bool) (val bool) {
	s := m.Get(key)
	if len(s) == 0 {
		return def
	}

	b, err := strconv.ParseBool(m.Get(key))
	if err != nil {
		return def
	}

	return b
}
