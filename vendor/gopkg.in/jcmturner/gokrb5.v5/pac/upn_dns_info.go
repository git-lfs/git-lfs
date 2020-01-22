package pac

import (
	"encoding/binary"
	"sort"

	"gopkg.in/jcmturner/rpc.v0/ndr"
)

// UPNDNSInfo implements https://msdn.microsoft.com/en-us/library/dd240468.aspx
type UPNDNSInfo struct {
	UPNLength           uint16 // An unsigned 16-bit integer in little-endian format that specifies the length, in bytes, of the UPN field.
	UPNOffset           uint16 // An unsigned 16-bit integer in little-endian format that contains the offset to the beginning of the buffer, in bytes, from the beginning of the UPN_DNS_INFO structure.
	DNSDomainNameLength uint16
	DNSDomainNameOffset uint16
	Flags               uint32
	UPN                 string
	DNSDomain           string
}

const (
	upnNoUPNAttr = 31 // The user account object does not have the userPrincipalName attribute ([MS-ADA3] section 2.349) set. A UPN constructed by concatenating the user name with the DNS domain name of the account domain is provided.
)

// Unmarshal bytes into the UPN_DNSInfo struct
func (k *UPNDNSInfo) Unmarshal(b []byte) error {
	//The UPN_DNS_INFO structure is a simple structure that is not NDR-encoded.
	var p int
	var e binary.ByteOrder = binary.LittleEndian

	k.UPNLength = ndr.ReadUint16(&b, &p, &e)
	k.UPNOffset = ndr.ReadUint16(&b, &p, &e)
	k.DNSDomainNameLength = ndr.ReadUint16(&b, &p, &e)
	k.DNSDomainNameOffset = ndr.ReadUint16(&b, &p, &e)
	k.Flags = ndr.ReadUint32(&b, &p, &e)
	ub := b[k.UPNOffset : k.UPNOffset+k.UPNLength]
	db := b[k.DNSDomainNameOffset : k.DNSDomainNameOffset+k.DNSDomainNameLength]

	u := make([]rune, k.UPNLength/2, k.UPNLength/2)
	for i := 0; i < len(u); i++ {
		q := i * 2
		u[i] = rune(ndr.ReadUint16(&ub, &q, &e))
	}
	k.UPN = string(u)
	d := make([]rune, k.DNSDomainNameLength/2, k.DNSDomainNameLength/2)
	for i := 0; i < len(d); i++ {
		q := i * 2
		d[i] = rune(ndr.ReadUint16(&db, &q, &e))
	}
	k.DNSDomain = string(d)

	l := []int{
		p,
		int(k.UPNOffset + k.UPNLength),
		int(k.DNSDomainNameOffset + k.DNSDomainNameLength),
	}
	sort.Ints(l)
	//Check that there is only zero padding left
	for _, v := range b[l[2]:] {
		if v != 0 {
			return ndr.Malformed{EText: "non-zero padding left over at end of data stream."}
		}
	}

	return nil
}
