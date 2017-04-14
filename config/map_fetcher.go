package config

// mapFetcher provides an implementation of the Fetcher interface by wrapping
// the `map[string]string` type.
type mapFetcher map[string][]string

func UniqMapFetcher(m map[string]string) Fetcher {
	multi := make(map[string][]string, len(m))
	for k, v := range m {
		multi[k] = []string{v}
	}

	return MapFetcher(multi)
}

func MapFetcher(m map[string][]string) Fetcher {
	return mapFetcher(m)
}

// Get implements the func `Fetcher.Get`.
func (m mapFetcher) Get(key string) (val string, ok bool) {
	all := m.GetAll(key)

	if len(all) == 0 {
		return "", false
	}
	return all[len(all)-1], true
}

// Get implements the func `Fetcher.GetAll`.
func (m mapFetcher) GetAll(key string) []string {
	return m[key]
}

func (m mapFetcher) All() map[string][]string {
	newmap := make(map[string][]string)
	for key, values := range m {
		for _, value := range values {
			newmap[key] = append(newmap[key], value)
		}
	}
	return newmap
}
