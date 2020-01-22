package mstypes

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"gopkg.in/jcmturner/rpc.v0/ndr"
)

// RPCSID implements https://msdn.microsoft.com/en-us/library/cc230364.aspx
type RPCSID struct {
	Revision            uint8                     // An 8-bit unsigned integer that specifies the revision level of the SID. This value MUST be set to 0x01.
	SubAuthorityCount   uint8                     // An 8-bit unsigned integer that specifies the number of elements in the SubAuthority array. The maximum number of elements allowed is 15.
	IdentifierAuthority RPCSIDIdentifierAuthority // An RPC_SID_IDENTIFIER_AUTHORITY structure that indicates the authority under which the SID was created. It describes the entity that created the SID. The Identifier Authority value {0,0,0,0,0,5} denotes SIDs created by the NT SID authority.
	SubAuthority        []uint32                  // A variable length array of unsigned 32-bit integers that uniquely identifies a principal relative to the IdentifierAuthority. Its length is determined by SubAuthorityCount.
}

// RPCSIDIdentifierAuthority implements https://msdn.microsoft.com/en-us/library/cc230372.aspx
type RPCSIDIdentifierAuthority struct {
	Value []byte // 6 bytes
}

// ReadRPCSID reads a RPC_SID from the bytes slice.
func ReadRPCSID(b *[]byte, p *int, e *binary.ByteOrder) (RPCSID, error) {
	size := int(ndr.ReadUint32(b, p, e)) // This is part of the NDR encoding rather than the data type.
	r := ndr.ReadUint8(b, p)
	if r != uint8(1) {
		return RPCSID{}, ndr.Malformed{EText: fmt.Sprintf("SID revision value read as %d when it must be 1", r)}
	}
	c := ndr.ReadUint8(b, p)
	a := ReadRPCSIDIdentifierAuthority(b, p, e)
	s := make([]uint32, c, c)
	if size != len(s) {
		return RPCSID{}, ndr.Malformed{EText: fmt.Sprintf("Number of elements (%d) within SID in the byte stream does not equal the SubAuthorityCount (%d)", size, c)}
	}
	for i := 0; i < len(s); i++ {
		s[i] = ndr.ReadUint32(b, p, e)
	}
	return RPCSID{
		Revision:            r,
		SubAuthorityCount:   c,
		IdentifierAuthority: a,
		SubAuthority:        s,
	}, nil
}

// ReadRPCSIDIdentifierAuthority reads a RPC_SIDIdentifierAuthority from the bytes slice.
func ReadRPCSIDIdentifierAuthority(b *[]byte, p *int, e *binary.ByteOrder) RPCSIDIdentifierAuthority {
	return RPCSIDIdentifierAuthority{
		Value: ndr.ReadBytes(b, p, 6, e),
	}
}

// ToString returns the string representation of the RPC_SID.
func (s *RPCSID) ToString() string {
	var str string
	b := append(make([]byte, 2, 2), s.IdentifierAuthority.Value...)
	// For a strange reason this is read big endian: https://msdn.microsoft.com/en-us/library/dd302645.aspx
	i := binary.BigEndian.Uint64(b)
	if i >= 4294967296 {
		str = fmt.Sprintf("S-1-0x%s", hex.EncodeToString(s.IdentifierAuthority.Value))
	} else {
		str = fmt.Sprintf("S-1-%d", i)
	}
	for _, sub := range s.SubAuthority {
		str = fmt.Sprintf("%s-%d", str, sub)
	}
	return str
}
