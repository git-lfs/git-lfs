package config

// Fetcher provides an interface to get typed information out of a configuration
// "source". These sources could be the OS enviornment, a .gitconfig, or even
// just a `map`.
type Fetcher interface {
	// Get returns the string value assosicated with a given key, or an
	// empty string if none exists.
	Get(key string) (val string)
}
