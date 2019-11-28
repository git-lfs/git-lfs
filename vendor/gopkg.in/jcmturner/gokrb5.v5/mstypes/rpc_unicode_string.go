package mstypes

import (
	"encoding/binary"

	"gopkg.in/jcmturner/rpc.v0/ndr"
)

// RPCUnicodeString implements https://msdn.microsoft.com/en-us/library/cc230365.aspx
type RPCUnicodeString struct {
	Length        uint16 // The length, in bytes, of the string pointed to by the Buffer member, not including the terminating null character if any. The length MUST be a multiple of 2. The length SHOULD equal the entire size of the Buffer, in which case there is no terminating null character. Any method that accesses this structure MUST use the Length specified instead of relying on the presence or absence of a null character.
	MaximumLength uint16 // The maximum size, in bytes, of the string pointed to by Buffer. The size MUST be a multiple of 2. If not, the size MUST be decremented by 1 prior to use. This value MUST not be less than Length.
	BufferPrt     uint32 // A pointer to a string buffer. If MaximumLength is greater than zero, the buffer MUST contain a non-null value.
	Value         string
}

// ReadRPCUnicodeString reads a RPCUnicodeString from the bytes slice.
func ReadRPCUnicodeString(b *[]byte, p *int, e *binary.ByteOrder) (RPCUnicodeString, error) {
	l := ndr.ReadUint16(b, p, e)
	ml := ndr.ReadUint16(b, p, e)
	if ml < l || l%2 != 0 || ml%2 != 0 {
		return RPCUnicodeString{}, ndr.Malformed{EText: "Invalid data for RPC_UNICODE_STRING"}
	}
	ptr := ndr.ReadUint32(b, p, e)
	return RPCUnicodeString{
		Length:        l,
		MaximumLength: ml,
		BufferPrt:     ptr,
	}, nil
}

// UnmarshalString populates a golang string into the RPCUnicodeString struct.
func (s *RPCUnicodeString) UnmarshalString(b *[]byte, p *int, e *binary.ByteOrder) (err error) {
	s.Value, err = ndr.ReadConformantVaryingString(b, p, e)
	return
}
