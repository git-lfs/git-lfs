package scanner

import "sync"

type NameCache struct {
	m  map[string]string
	mu *sync.Mutex
}

func NewNameCache(m map[string]string) *NameCache {
	return &NameCache{
		m:  m,
		mu: &sync.Mutex{},
	}
}

func (c *NameCache) GetName(sha string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	rev, ok := c.m[sha]
	return rev, ok
}

func (c *NameCache) Cache(sha, name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.m[sha] = name
}
