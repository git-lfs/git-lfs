package endpoint

import (
	"fmt"
	"strings"
)

type Access string

const (
	NoneAccess    Access = "none"
	BasicAccess   Access = "basic"
	PrivateAccess Access = "private"
	NTLMAccess    Access = "ntlm"
)

func (c *Config) AccessFor(e Endpoint) Access {
	if c.git == nil {
		return NoneAccess
	}

	key := fmt.Sprintf("lfs.%s.access", e.Url)
	if v, ok := c.git.Get(key); ok && len(v) > 0 {
		lower := Access(strings.ToLower(v))
		if lower == PrivateAccess {
			return BasicAccess
		}
		return lower
	}
	return NoneAccess
}
