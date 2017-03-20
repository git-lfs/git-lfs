package config

// Fetcher provides an interface to get typed information out of a configuration
// "source". These sources could be the OS enviornment, a .gitconfig, or even
// just a `map`.
type Fetcher interface {
	// Get returns the string value associated with a given key and a bool
	// determining if the key exists.
	Get(key string) (val string, ok bool)

	// All returns a copy of all the key/value pairs for the current environment.
	All() map[string]string
}
