package scanner

import "sync"

// NameCache is a goroutine-safe cache of the SHA1 hash of a file, to its name
// on disk.
type NameCache struct {
	// m maps the SHA1 hash of a file to its name on disk.
	m map[string]string
	// mu guards m.
	mu *sync.Mutex
}

// NewNameCache instantiates a new `*NameCache` using the given map, `m`.
//
// NOTE: passing `m` in from the call-site is only a temporary measure, and
// exists so old references into ScanOpts can still remain valid.
func NewNameCache() *NameCache {
	return &NameCache{
		m:  make(map[string]string),
		mu: &sync.Mutex{},
	}
}

// GetName returns the name of the file who's contents has the SHA1 hash of
// "sha". If no name was found matching the given "sha", then an empty string
// and the value false are returned, true otherwise.
func (c *NameCache) GetName(sha string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	rev, ok := c.m[sha]
	return rev, ok
}

// Cache caches the given SHA1, "sha", against the given name, "name",
// overwriting pre-existing cache entries, if any.
func (c *NameCache) Cache(sha, name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.m[sha] = name
}
