package pac

import (
	"fmt"

	"gopkg.in/jcmturner/gokrb5.v5/mstypes"
	"gopkg.in/jcmturner/rpc.v0/ndr"
)

// Claims reference: https://msdn.microsoft.com/en-us/library/hh553895.aspx

// ClientClaimsInfo implements https://msdn.microsoft.com/en-us/library/hh536365.aspx
type ClientClaimsInfo struct {
	Claims mstypes.ClaimsSetMetadata
}

// Unmarshal bytes into the ClientClaimsInfo struct
func (k *ClientClaimsInfo) Unmarshal(b []byte) error {
	ch, _, p, err := ndr.ReadHeaders(&b)
	if err != nil {
		return fmt.Errorf("error parsing byte stream headers of CLIENT_CLAIMS_INFO: %v", err)
	}
	e := &ch.Endianness
	//The next 4 bytes are an RPC unique pointer referent. We just skip these
	p += 4

	k.Claims, err = mstypes.ReadClaimsSetMetadata(&b, &p, e)
	if err != nil {
		return err
	}

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
