package scanner

import "sync"

type RevCache struct {
	m  map[string]string
	mu *sync.Mutex
}

func NewRevCache(m map[string]string) *RevCache {
	return &RevCache{
		m:  m,
		mu: &sync.Mutex{},
	}
}

func (c *RevCache) GetName(sha string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	rev, ok := c.m[sha]
	return rev, ok
}

func (c *RevCache) Cache(sha, name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.m[sha] = name
}
