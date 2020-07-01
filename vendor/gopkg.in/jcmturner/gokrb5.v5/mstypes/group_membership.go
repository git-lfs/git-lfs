package mstypes

import (
	"encoding/binary"

	"gopkg.in/jcmturner/rpc.v0/ndr"
)

// GroupMembership implements https://msdn.microsoft.com/en-us/library/cc237945.aspx
// RelativeID : A 32-bit unsigned integer that contains the RID of a particular group.
// The possible values for the Attributes flags are identical to those specified in KERB_SID_AND_ATTRIBUTES
type GroupMembership struct {
	RelativeID uint32
	Attributes uint32
}

// ReadGroupMembership reads a GroupMembership from the bytes slice.
func ReadGroupMembership(b *[]byte, p *int, e *binary.ByteOrder) GroupMembership {
	r := ndr.ReadUint32(b, p, e)
	a := ndr.ReadUint32(b, p, e)
	return GroupMembership{
		RelativeID: r,
		Attributes: a,
	}
}

// DomainGroupMembership implements https://msdn.microsoft.com/en-us/library/hh536344.aspx
// DomainId: A SID structure that contains the SID for the domain.This member is used in conjunction with the GroupIds members to create group SIDs for the device.
// GroupCount: A 32-bit unsigned integer that contains the number of groups within the domain to which the account belongs.
// GroupIds: A pointer to a list of GROUP_MEMBERSHIP structures that contain the groups to which the account belongs in the domain. The number of groups in this list MUST be equal to GroupCount.
type DomainGroupMembership struct {
	DomainID   RPCSID
	GroupCount uint32
	GroupIDs   []GroupMembership // Size is value of GroupCount
}

// ReadDomainGroupMembership reads a DomainGroupMembership from the bytes slice.
func ReadDomainGroupMembership(b *[]byte, p *int, e *binary.ByteOrder) (DomainGroupMembership, error) {
	d, err := ReadRPCSID(b, p, e)
	if err != nil {
		return DomainGroupMembership{}, err
	}
	c := ndr.ReadUint32(b, p, e)
	g := make([]GroupMembership, c, c)
	for i := range g {
		g[i] = ReadGroupMembership(b, p, e)
	}
	return DomainGroupMembership{
		DomainID:   d,
		GroupCount: c,
		GroupIDs:   g,
	}, nil
}
