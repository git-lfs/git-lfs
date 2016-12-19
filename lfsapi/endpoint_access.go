package lfsapi

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

func (e *endpointGitFinder) AccessFor(ep Endpoint) Access {
	if e.git == nil {
		return NoneAccess
	}

	key := fmt.Sprintf("lfs.%s.access", ep.Url)
	if v, ok := e.git.Get(key); ok && len(v) > 0 {
		lower := Access(strings.ToLower(v))
		if lower == PrivateAccess {
			return BasicAccess
		}
		return lower
	}
	return NoneAccess
}
