package pack

type ChainSimple struct {
	X   []byte
	Err error
}

func (c *ChainSimple) Unpack() ([]byte, error) {
	return c.X, c.Err
}

func (c *ChainSimple) Type() PackedObjectType { return TypeNone }
