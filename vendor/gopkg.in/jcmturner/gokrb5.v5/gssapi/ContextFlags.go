package gssapi

import "github.com/jcmturner/gofork/encoding/asn1"

/*
ContextFlags ::= BIT STRING {
  delegFlag       (0),
  mutualFlag      (1),
  replayFlag      (2),
  sequenceFlag    (3),
  anonFlag        (4),
  confFlag        (5),
  integFlag       (6)
} (SIZE (32))
*/

const (
	delegFlag    = 0
	mutualFlag   = 1
	replayFlag   = 2
	sequenceFlag = 3
	anonFlag     = 4
	confFlag     = 5
	integFlag    = 6
)

// ContextFlags flags for GSSAPI
type ContextFlags asn1.BitString

// NewContextFlags creates a new ContextFlags instance.
func NewContextFlags() ContextFlags {
	var c ContextFlags
	c.BitLength = 32
	c.Bytes = make([]byte, 4)
	return c
}
