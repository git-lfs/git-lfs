package pac

import (
	"encoding/binary"

	"gopkg.in/jcmturner/rpc.v0/ndr"
)

const (
	ulTypeKerbValidationInfo     = 1
	ulTypeCredentials            = 2
	ulTypePACServerSignatureData = 6
	ulTypePACKDCSignatureData    = 7
	ulTypePACClientInfo          = 10
	ulTypeS4UDelegationInfo      = 11
	ulTypeUPNDNSInfo             = 12
	ulTypePACClientClaimsInfo    = 13
	ulTypePACDeviceInfo          = 14
	ulTypePACDeviceClaimsInfo    = 15
)

// InfoBuffer implements the PAC Info Buffer: https://msdn.microsoft.com/en-us/library/cc237954.aspx
type InfoBuffer struct {
	ULType       uint32 // A 32-bit unsigned integer in little-endian format that describes the type of data present in the buffer contained at Offset.
	CBBufferSize uint32 // A 32-bit unsigned integer in little-endian format that contains the size, in bytes, of the buffer in the PAC located at Offset.
	Offset       uint64 // A 64-bit unsigned integer in little-endian format that contains the offset to the beginning of the buffer, in bytes, from the beginning of the PACTYPE structure. The data offset MUST be a multiple of eight. The following sections specify the format of each type of element.
}

// ReadPACInfoBuffer reads a InfoBuffer from the byte slice.
func ReadPACInfoBuffer(b *[]byte, p *int, e *binary.ByteOrder) InfoBuffer {
	u := ndr.ReadUint32(b, p, e)
	s := ndr.ReadUint32(b, p, e)
	o := ndr.ReadUint64(b, p, e)
	return InfoBuffer{
		ULType:       u,
		CBBufferSize: s,
		Offset:       o,
	}
}
