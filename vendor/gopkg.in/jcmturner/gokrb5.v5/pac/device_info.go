package pac

import (
	"fmt"

	"gopkg.in/jcmturner/gokrb5.v5/mstypes"
	"gopkg.in/jcmturner/rpc.v0/ndr"
)

// DeviceInfo implements https://msdn.microsoft.com/en-us/library/hh536402.aspx
type DeviceInfo struct {
	UserID            uint32                          // A 32-bit unsigned integer that contains the RID of the account. If the UserId member equals 0x00000000, the first group SID in this member is the SID for this account.
	PrimaryGroupID    uint32                          // A 32-bit unsigned integer that contains the RID for the primary group to which this account belongs.
	AccountDomainID   mstypes.RPCSID                  // A SID structure that contains the SID for the domain of the account.This member is used in conjunction with the UserId, and GroupIds members to create the user and group SIDs for the client.
	AccountGroupCount uint32                          // A 32-bit unsigned integer that contains the number of groups within the account domain to which the account belongs
	AccountGroupIDs   []mstypes.GroupMembership       // A pointer to a list of GROUP_MEMBERSHIP (section 2.2.2) structures that contains the groups to which the account belongs in the account domain. The number of groups in this list MUST be equal to GroupCount.
	SIDCount          uint32                          // A 32-bit unsigned integer that contains the total number of SIDs present in the ExtraSids member.
	ExtraSIDs         []mstypes.KerbSidAndAttributes  // A pointer to a list of KERB_SID_AND_ATTRIBUTES structures that contain a list of SIDs corresponding to groups not in domains. If the UserId member equals 0x00000000, the first group SID in this member is the SID for this account.
	DomainGroupCount  uint32                          // A 32-bit unsigned integer that contains the number of domains with groups to which the account belongs.
	DomainGroup       []mstypes.DomainGroupMembership // A pointer to a list of DOMAIN_GROUP_MEMBERSHIP structures (section 2.2.3) that contains the domains to which the account belongs to a group. The number of sets in this list MUST be equal to DomainCount.
}

// Unmarshal bytes into the DeviceInfo struct
func (k *DeviceInfo) Unmarshal(b []byte) error {
	ch, _, p, err := ndr.ReadHeaders(&b)
	if err != nil {
		return fmt.Errorf("error parsing byte stream headers: %v", err)
	}
	e := &ch.Endianness

	//The next 4 bytes are an RPC unique pointer referent. We just skip these
	p += 4

	k.UserID = ndr.ReadUint32(&b, &p, e)
	k.PrimaryGroupID = ndr.ReadUint32(&b, &p, e)
	k.AccountDomainID, err = mstypes.ReadRPCSID(&b, &p, e)
	if err != nil {
		return err
	}
	k.AccountGroupCount = ndr.ReadUint32(&b, &p, e)
	if k.AccountGroupCount > 0 {
		ag := make([]mstypes.GroupMembership, k.AccountGroupCount, k.AccountGroupCount)
		for i := range ag {
			ag[i] = mstypes.ReadGroupMembership(&b, &p, e)
		}
		k.AccountGroupIDs = ag
	}

	k.SIDCount = ndr.ReadUint32(&b, &p, e)
	var ah ndr.ConformantArrayHeader
	if k.SIDCount > 0 {
		ah, err = ndr.ReadUniDimensionalConformantArrayHeader(&b, &p, e)
		if ah.MaxCount != int(k.SIDCount) {
			return fmt.Errorf("error with size of ExtraSIDs list. expected: %d, Actual: %d", k.SIDCount, ah.MaxCount)
		}
		es := make([]mstypes.KerbSidAndAttributes, k.SIDCount, k.SIDCount)
		attr := make([]uint32, k.SIDCount, k.SIDCount)
		ptr := make([]uint32, k.SIDCount, k.SIDCount)
		for i := range attr {
			ptr[i] = ndr.ReadUint32(&b, &p, e)
			attr[i] = ndr.ReadUint32(&b, &p, e)
		}
		for i := range es {
			if ptr[i] != 0 {
				s, err := mstypes.ReadRPCSID(&b, &p, e)
				es[i] = mstypes.KerbSidAndAttributes{SID: s, Attributes: attr[i]}
				if err != nil {
					return ndr.Malformed{EText: fmt.Sprintf("could not read ExtraSIDs: %v", err)}
				}
			}
		}
		k.ExtraSIDs = es
	}

	k.DomainGroupCount = ndr.ReadUint32(&b, &p, e)
	if k.DomainGroupCount > 0 {
		dg := make([]mstypes.DomainGroupMembership, k.DomainGroupCount, k.DomainGroupCount)
		for i := range dg {
			dg[i], _ = mstypes.ReadDomainGroupMembership(&b, &p, e)
		}
		k.DomainGroup = dg
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
