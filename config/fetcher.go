package config

// Fetcher provides an interface to get typed information out of a configuration
// "source". These sources could be the OS enviornment, a .gitconfig, or even
// just a `map`.
type Fetcher interface {
	// Get returns the string value associated with a given key and a bool
	// determining if the key exists.
	//
	// If multiple entries match the given key, the first one will be
	// returned.
	Get(key string) (val string, ok bool)

	// GetAll returns the a set of string values associated with a given
	// key. If no entries matched the given key, an empty slice will be
	// returned instead.
	GetAll(key string) (vals []string)

	// All returns a copy of all the key/value pairs for the current
	// environment.
	All() map[string][]string
}
