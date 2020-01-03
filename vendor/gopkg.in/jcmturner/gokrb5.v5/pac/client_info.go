package pac

import (
	"encoding/binary"

	"gopkg.in/jcmturner/gokrb5.v5/mstypes"
	"gopkg.in/jcmturner/rpc.v0/ndr"
)

// ClientInfo implements https://msdn.microsoft.com/en-us/library/cc237951.aspx
type ClientInfo struct {
	ClientID   mstypes.FileTime // A FILETIME structure in little-endian format that contains the Kerberos initial ticket-granting ticket TGT authentication time
	NameLength uint16           // An unsigned 16-bit integer in little-endian format that specifies the length, in bytes, of the Name field.
	Name       string           // An array of 16-bit Unicode characters in little-endian format that contains the client's account name.
}

// Unmarshal bytes into the ClientInfo struct
func (k *ClientInfo) Unmarshal(b []byte) error {
	//The PAC_CLIENT_INFO structure is a simple structure that is not NDR-encoded.
	var p int
	var e binary.ByteOrder = binary.LittleEndian

	k.ClientID = mstypes.ReadFileTime(&b, &p, &e)
	k.NameLength = ndr.ReadUint16(&b, &p, &e)
	if len(b[p:]) < int(k.NameLength) {
		return ndr.Malformed{EText: "PAC ClientInfo length truncated"}
	}
	k.Name = ndr.ReadUTF16String(int(k.NameLength), &b, &p, &e)

	//Check that there is only zero padding left
	if len(b) >= p {
		for _, v := range b[p:] {
			if v != 0 {
				return ndr.Malformed{EText: "non-zero padding left over at end of data stream"}
			}
		}
	}

	return nil
}
