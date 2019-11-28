// Package pac implements Microsoft Privilege Attribute Certificate (PAC) processing.
package pac

import (
	"errors"
	"fmt"

	"gopkg.in/jcmturner/gokrb5.v5/mstypes"
	"gopkg.in/jcmturner/rpc.v0/ndr"
)

// KERB_VALIDATION_INFO flags.
const (
	USERFLAG_GUEST                                    = 31 // Authentication was done via the GUEST account; no password was used.
	USERFLAG_NO_ENCRYPTION_AVAILABLE                  = 30 // No encryption is available.
	USERFLAG_LAN_MANAGER_KEY                          = 28 // LAN Manager key was used for authentication.
	USERFLAG_SUB_AUTH                                 = 25 // Sub-authentication used; session key came from the sub-authentication package.
	USERFLAG_EXTRA_SIDS                               = 26 // Indicates that the ExtraSids field is populated and contains additional SIDs.
	USERFLAG_MACHINE_ACCOUNT                          = 24 // Indicates that the account is a machine account.
	USERFLAG_DC_NTLM2                                 = 23 // Indicates that the domain controller understands NTLMv2.
	USERFLAG_RESOURCE_GROUPIDS                        = 22 // Indicates that the ResourceGroupIds field is populated.
	USERFLAG_PROFILEPATH                              = 21 // Indicates that ProfilePath is populated.
	USERFLAG_NTLM2_NTCHALLENGERESP                    = 20 // The NTLMv2 response from the NtChallengeResponseFields ([MS-NLMP] section 2.2.1.3) was used for authentication and session key generation.
	USERFLAG_LM2_LMCHALLENGERESP                      = 19 // The LMv2 response from the LmChallengeResponseFields ([MS-NLMP] section 2.2.1.3) was used for authentication and session key generation.
	USERFLAG_AUTH_LMCHALLENGERESP_KEY_NTCHALLENGERESP = 18 // The LMv2 response from the LmChallengeResponseFields ([MS-NLMP] section 2.2.1.3) was used for authentication and the NTLMv2 response from the NtChallengeResponseFields ([MS-NLMP] section 2.2.1.3) was used session key generation.
)

// KerbValidationInfo implement https://msdn.microsoft.com/en-us/library/cc237948.aspx
// The KERB_VALIDATION_INFO structure defines the user's logon and authorization information
// provided by the DC. The KERB_VALIDATION_INFO structure is a subset of the
// NETLOGON_VALIDATION_SAM_INFO4 structure ([MS-NRPC] section 2.2.1.4.13).
// It is a subset due to historical reasons and to the use of the common Active Directory to generate this information.
// The KERB_VALIDATION_INFO structure is marshaled by RPC [MS-RPCE].
type KerbValidationInfo struct {
	LogOnTime               mstypes.FileTime
	LogOffTime              mstypes.FileTime
	KickOffTime             mstypes.FileTime
	PasswordLastSet         mstypes.FileTime
	PasswordCanChange       mstypes.FileTime
	PasswordMustChange      mstypes.FileTime
	EffectiveName           mstypes.RPCUnicodeString
	FullName                mstypes.RPCUnicodeString
	LogonScript             mstypes.RPCUnicodeString
	ProfilePath             mstypes.RPCUnicodeString
	HomeDirectory           mstypes.RPCUnicodeString
	HomeDirectoryDrive      mstypes.RPCUnicodeString
	LogonCount              uint16
	BadPasswordCount        uint16
	UserID                  uint32
	PrimaryGroupID          uint32
	GroupCount              uint32
	pGroupIDs               uint32
	GroupIDs                []mstypes.GroupMembership
	UserFlags               uint32
	UserSessionKey          mstypes.UserSessionKey
	LogonServer             mstypes.RPCUnicodeString
	LogonDomainName         mstypes.RPCUnicodeString
	pLogonDomainID          uint32
	LogonDomainID           mstypes.RPCSID
	Reserved1               []uint32 // Has 2 elements
	UserAccountControl      uint32
	SubAuthStatus           uint32
	LastSuccessfulILogon    mstypes.FileTime
	LastFailedILogon        mstypes.FileTime
	FailedILogonCount       uint32
	Reserved3               uint32
	SIDCount                uint32
	pExtraSIDs              uint32
	ExtraSIDs               []mstypes.KerbSidAndAttributes
	pResourceGroupDomainSID uint32
	ResourceGroupDomainSID  mstypes.RPCSID
	ResourceGroupCount      uint32
	pResourceGroupIDs       uint32
	ResourceGroupIDs        []mstypes.GroupMembership
}

// Unmarshal bytes into the DeviceInfo struct
func (k *KerbValidationInfo) Unmarshal(b []byte) (err error) {
	ch, _, p, err := ndr.ReadHeaders(&b)
	if err != nil {
		return fmt.Errorf("error parsing byte stream headers: %v", err)
	}
	e := &ch.Endianness

	//The next 4 bytes are an RPC unique pointer referent. We just skip these
	p += 4

	k.LogOnTime = mstypes.ReadFileTime(&b, &p, e)
	k.LogOffTime = mstypes.ReadFileTime(&b, &p, e)
	k.KickOffTime = mstypes.ReadFileTime(&b, &p, e)
	k.PasswordLastSet = mstypes.ReadFileTime(&b, &p, e)
	k.PasswordCanChange = mstypes.ReadFileTime(&b, &p, e)
	k.PasswordMustChange = mstypes.ReadFileTime(&b, &p, e)

	if k.EffectiveName, err = mstypes.ReadRPCUnicodeString(&b, &p, e); err != nil {
		return
	}
	if k.FullName, err = mstypes.ReadRPCUnicodeString(&b, &p, e); err != nil {
		return
	}
	if k.LogonScript, err = mstypes.ReadRPCUnicodeString(&b, &p, e); err != nil {
		return
	}
	if k.ProfilePath, err = mstypes.ReadRPCUnicodeString(&b, &p, e); err != nil {
		return
	}
	if k.HomeDirectory, err = mstypes.ReadRPCUnicodeString(&b, &p, e); err != nil {
		return
	}
	if k.HomeDirectoryDrive, err = mstypes.ReadRPCUnicodeString(&b, &p, e); err != nil {
		return
	}

	k.LogonCount = ndr.ReadUint16(&b, &p, e)
	k.BadPasswordCount = ndr.ReadUint16(&b, &p, e)
	k.UserID = ndr.ReadUint32(&b, &p, e)
	k.PrimaryGroupID = ndr.ReadUint32(&b, &p, e)
	k.GroupCount = ndr.ReadUint32(&b, &p, e)
	k.pGroupIDs = ndr.ReadUint32(&b, &p, e)

	k.UserFlags = ndr.ReadUint32(&b, &p, e)
	k.UserSessionKey = mstypes.ReadUserSessionKey(&b, &p, e)

	if k.LogonServer, err = mstypes.ReadRPCUnicodeString(&b, &p, e); err != nil {
		return
	}
	if k.LogonDomainName, err = mstypes.ReadRPCUnicodeString(&b, &p, e); err != nil {
		return
	}

	k.pLogonDomainID = ndr.ReadUint32(&b, &p, e)

	k.Reserved1 = []uint32{
		ndr.ReadUint32(&b, &p, e),
		ndr.ReadUint32(&b, &p, e),
	}

	k.UserAccountControl = ndr.ReadUint32(&b, &p, e)
	k.SubAuthStatus = ndr.ReadUint32(&b, &p, e)
	k.LastSuccessfulILogon = mstypes.ReadFileTime(&b, &p, e)
	k.LastFailedILogon = mstypes.ReadFileTime(&b, &p, e)
	k.FailedILogonCount = ndr.ReadUint32(&b, &p, e)
	k.Reserved3 = ndr.ReadUint32(&b, &p, e)

	k.SIDCount = ndr.ReadUint32(&b, &p, e)
	k.pExtraSIDs = ndr.ReadUint32(&b, &p, e)

	k.pResourceGroupDomainSID = ndr.ReadUint32(&b, &p, e)
	k.ResourceGroupCount = ndr.ReadUint32(&b, &p, e)
	k.pResourceGroupIDs = ndr.ReadUint32(&b, &p, e)

	// Populate pointers
	if err = k.EffectiveName.UnmarshalString(&b, &p, e); err != nil {
		return
	}
	if err = k.FullName.UnmarshalString(&b, &p, e); err != nil {
		return
	}
	if err = k.LogonScript.UnmarshalString(&b, &p, e); err != nil {
		return
	}
	if err = k.ProfilePath.UnmarshalString(&b, &p, e); err != nil {
		return
	}
	if err = k.HomeDirectory.UnmarshalString(&b, &p, e); err != nil {
		return
	}
	if err = k.HomeDirectoryDrive.UnmarshalString(&b, &p, e); err != nil {
		return
	}
	var ah ndr.ConformantArrayHeader
	if k.GroupCount > 0 {
		ah, err = ndr.ReadUniDimensionalConformantArrayHeader(&b, &p, e)
		if err != nil {
			return
		}
		if ah.MaxCount != int(k.GroupCount) {
			err = errors.New("error with size of group list")
			return
		}
		g := make([]mstypes.GroupMembership, k.GroupCount, k.GroupCount)
		for i := range g {
			g[i] = mstypes.ReadGroupMembership(&b, &p, e)
		}
		k.GroupIDs = g
	}

	if err = k.LogonServer.UnmarshalString(&b, &p, e); err != nil {
		return
	}
	if err = k.LogonDomainName.UnmarshalString(&b, &p, e); err != nil {
		return
	}

	if k.pLogonDomainID != 0 {
		k.LogonDomainID, err = mstypes.ReadRPCSID(&b, &p, e)
		if err != nil {
			return fmt.Errorf("error reading LogonDomainID: %v", err)
		}
	}

	if k.SIDCount > 0 {
		ah, err = ndr.ReadUniDimensionalConformantArrayHeader(&b, &p, e)
		if err != nil {
			return
		}
		if ah.MaxCount != int(k.SIDCount) {
			return fmt.Errorf("error with size of ExtraSIDs list. Expected: %d, Actual: %d", k.SIDCount, ah.MaxCount)
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

	if k.pResourceGroupDomainSID != 0 {
		k.ResourceGroupDomainSID, err = mstypes.ReadRPCSID(&b, &p, e)
		if err != nil {
			return err
		}
	}

	if k.ResourceGroupCount > 0 {
		ah, err = ndr.ReadUniDimensionalConformantArrayHeader(&b, &p, e)
		if err != nil {
			return
		}
		if ah.MaxCount != int(k.ResourceGroupCount) {
			return fmt.Errorf("error with size of ResourceGroup list. Expected: %d, Actual: %d", k.ResourceGroupCount, ah.MaxCount)
		}
		g := make([]mstypes.GroupMembership, k.ResourceGroupCount, k.ResourceGroupCount)
		for i := range g {
			g[i] = mstypes.ReadGroupMembership(&b, &p, e)
		}
		k.ResourceGroupIDs = g
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

// GetGroupMembershipSIDs returns a slice of strings containing the group membership SIDs found in the PAC.
func (k *KerbValidationInfo) GetGroupMembershipSIDs() []string {
	var g []string
	lSID := k.LogonDomainID.ToString()
	for i := range k.GroupIDs {
		g = append(g, fmt.Sprintf("%s-%d", lSID, k.GroupIDs[i].RelativeID))
	}
	for _, s := range k.ExtraSIDs {
		var exists = false
		for _, es := range g {
			if es == s.SID.ToString() {
				exists = true
				break
			}
		}
		if !exists {
			g = append(g, s.SID.ToString())
		}
	}
	for _, r := range k.ResourceGroupIDs {
		var exists = false
		s := fmt.Sprintf("%s-%d", k.ResourceGroupDomainSID.ToString(), r.RelativeID)
		for _, es := range g {
			if es == s {
				exists = true
				break
			}
		}
		if !exists {
			g = append(g, s)
		}
	}
	return g
}
