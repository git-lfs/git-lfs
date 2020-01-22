package mstypes

import (
	"encoding/binary"

	"gopkg.in/jcmturner/rpc.v0/ndr"
)

// CypherBlock implements https://msdn.microsoft.com/en-us/library/cc237040.aspx
type CypherBlock struct {
	Data []byte // size = 8
}

// UserSessionKey implements https://msdn.microsoft.com/en-us/library/cc237080.aspx
type UserSessionKey struct {
	Data []CypherBlock // size = 2
}

// ReadUserSessionKey reads a UserSessionKey from the bytes slice.
func ReadUserSessionKey(b *[]byte, p *int, e *binary.ByteOrder) UserSessionKey {
	cb1 := CypherBlock{
		Data: ndr.ReadBytes(b, p, 8, e),
	}
	cb2 := CypherBlock{
		Data: ndr.ReadBytes(b, p, 8, e),
	}
	return UserSessionKey{
		Data: []CypherBlock{cb1, cb2},
	}
}
