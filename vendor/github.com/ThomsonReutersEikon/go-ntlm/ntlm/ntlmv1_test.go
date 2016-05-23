//Copyright 2013 Thomson Reuters Global Resources. BSD License please see License file for more information

package ntlm

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"testing"
)

func TestLMOWFv1(t *testing.T) {
	// Sample from MS-NLMP
	result, err := lmowfv1("Password")
	expected, _ := hex.DecodeString("e52cac67419a9a224a3b108f3fa6cb6d")
	if err != nil || !bytes.Equal(result, expected) {
		t.Errorf("LMNOWFv1 is not correct, got %s expected %s", hex.EncodeToString(result), "e52cac67419a9a224a3b108f3fa6cb6d")
	}
}

func TestNTOWFv1(t *testing.T) {
	// Sample from MS-NLMP
	result := ntowfv1("Password")
	expected, _ := hex.DecodeString("a4f49c406510bdcab6824ee7c30fd852")
	if !bytes.Equal(result, expected) {
		t.Error("NTOWFv1 is not correct")
	}
}

func checkV1Value(t *testing.T, name string, value []byte, expected string, err error) {
	if err != nil {
		t.Errorf("NTLMv1 %s received error: %s", name, err)
	} else {
		expectedBytes, _ := hex.DecodeString(expected)
		if !bytes.Equal(expectedBytes, value) {
			t.Errorf("NTLMv1 %s is not correct got %s expected %s", name, hex.EncodeToString(value), expected)
		}
	}
}

// There was an issue where all NTLMv1 authentications with extended session security
// would authenticate. This was due to a bug in the MS-NLMP docs. This tests for that issue
func TestNtlmV1ExtendedSessionSecurity(t *testing.T) {
	// NTLMv1 with extended session security
  challengeMessage := "TlRMTVNTUAACAAAAAAAAADgAAABVgphiRy3oSZvn1I4AAAAAAAAAAKIAogA4AAAABQEoCgAAAA8CAA4AUgBFAFUAVABFAFIAUwABABwAVQBLAEIAUAAtAEMAQgBUAFIATQBGAEUAMAA2AAQAFgBSAGUAdQB0AGUAcgBzAC4AbgBlAHQAAwA0AHUAawBiAHAALQBjAGIAdAByAG0AZgBlADAANgAuAFIAZQB1AHQAZQByAHMALgBuAGUAdAAFABYAUgBlAHUAdABlAHIAcwAuAG4AZQB0AAAAAAA="
  authenticateMessage := "TlRMTVNTUAADAAAAGAAYAJgAAAAYABgAsAAAAAAAAABIAAAAOgA6AEgAAAAWABYAggAAABAAEADIAAAAVYKYYgUCzg4AAAAPMQAwADAAMAAwADEALgB3AGMAcABAAHQAaABvAG0AcwBvAG4AcgBlAHUAdABlAHIAcwAuAGMAbwBtAE4AWQBDAFMATQBTAEcAOQA5ADAAOQBRWAK3h/TIywAAAAAAAAAAAAAAAAAAAAA3tp89kZU1hs1XZp7KTyGm3XsFAT9stEDW9YXDaeYVBmBcBb//2FOu"

	challengeData, _ := base64.StdEncoding.DecodeString(challengeMessage)
	c, _ := ParseChallengeMessage(challengeData)

  authenticateData, _ := base64.StdEncoding.DecodeString(authenticateMessage)
  msg, err := ParseAuthenticateMessage(authenticateData, 1)
	if err != nil {
		t.Errorf("Could not process authenticate message: %s", err)
	}

	context, err := CreateServerSession(Version1, ConnectionlessMode)
	if err != nil {
		t.Errorf("Could not create NTLMv1 session")
	}
	context.SetUserInfo("100001.wcp.thomsonreuters.com", "notmypass", "")
	context.SetServerChallenge(c.ServerChallenge)
	err = context.ProcessAuthenticateMessage(msg)
	if err == nil {
		t.Errorf("This message should have failed to authenticate, but it passed", err)
	}
}

func TestNtlmV1(t *testing.T) {
	flags := uint32(0)
	flags = NTLMSSP_NEGOTIATE_KEY_EXCH.Set(flags)
	flags = NTLMSSP_NEGOTIATE_56.Set(flags)
	flags = NTLMSSP_NEGOTIATE_128.Set(flags)
	flags = NTLMSSP_NEGOTIATE_VERSION.Set(flags)
	flags = NTLMSSP_TARGET_TYPE_SERVER.Set(flags)
	flags = NTLMSSP_NEGOTIATE_ALWAYS_SIGN.Set(flags)
	flags = NTLMSSP_NEGOTIATE_NTLM.Set(flags)
	flags = NTLMSSP_NEGOTIATE_SEAL.Set(flags)
	flags = NTLMSSP_NEGOTIATE_SIGN.Set(flags)
	flags = NTLM_NEGOTIATE_OEM.Set(flags)
	flags = NTLMSSP_NEGOTIATE_UNICODE.Set(flags)

	n := new(V1ClientSession)
	n.SetUserInfo("User", "Password", "Domain")
	n.NegotiateFlags = flags
	n.responseKeyNT, _ = hex.DecodeString("a4f49c406510bdcab6824ee7c30fd852")
	n.responseKeyLM, _ = hex.DecodeString("e52cac67419a9a224a3b108f3fa6cb6d")
	n.clientChallenge, _ = hex.DecodeString("aaaaaaaaaaaaaaaa")
	n.serverChallenge, _ = hex.DecodeString("0123456789abcdef")

	var err error
	// 4.2.2.1.3 Session Base Key and Key Exchange Key
	err = n.computeSessionBaseKey()
	checkV1Value(t, "sessionBaseKey", n.sessionBaseKey, "d87262b0cde4b1cb7499becccdf10784", err)
	err = n.computeKeyExchangeKey()
	checkV1Value(t, "keyExchangeKey", n.keyExchangeKey, "d87262b0cde4b1cb7499becccdf10784", err)

	// 4.2.2.2.1 NTLMv1 Response
	// NTChallengeResponse with With NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY not set
	err = n.computeExpectedResponses()
	checkV1Value(t, "NTChallengeResponse", n.ntChallengeResponse, "67c43011f30298a2ad35ece64f16331c44bdbed927841f94", err)
	// 4.2.2.2.2 LMv1 Response
	// The LmChallengeResponse is specified in section 3.3.1. With the NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY flag
	// not set and with the NoLMResponseNTLMv1 flag not set
	checkV1Value(t, "LMChallengeResponse", n.lmChallengeResponse, "98def7b87f88aa5dafe2df779688a172def11c7d5ccdef13", err)

	// If the NTLMSSP_NEGOTIATE_LM_KEY flag is set then the KeyExchangeKey is:
	n.NegotiateFlags = NTLMSSP_NEGOTIATE_LM_KEY.Set(n.NegotiateFlags)
	err = n.computeKeyExchangeKey()
	checkV1Value(t, "keyExchangeKey with NTLMSSP_NEGOTIATE_LM_KEY", n.keyExchangeKey, "b09e379f7fbecb1eaf0afdcb0383c8a0", err)
	n.NegotiateFlags = NTLMSSP_NEGOTIATE_LM_KEY.Unset(n.NegotiateFlags)

	// 4.2.2.2.3 Encrypted Session Key
	//n.randomSessionKey, _ = hex.DecodeString("55555555555555555555555555555555")

	// RC4 decryption of the EncryptedRandomSessionKey with the KeyExchange key
	//err = n.computeKeyExchangeKey()
	//n.encryptedRandomSessionKey, err = hex.DecodeString("518822b1b3f350c8958682ecbb3e3cb7")
	//err = n.computeExportedSessionKey()
	//checkV1Value(t, "ExportedSessionKey", n.exportedSessionKey, "55555555555555555555555555555555", err)

	// NTLMSSP_REQUEST_NON_NT_SESSION_KEY is set:
	n.NegotiateFlags = NTLMSSP_REQUEST_NON_NT_SESSION_KEY.Set(n.NegotiateFlags)
	err = n.computeKeyExchangeKey()
	//	n.encryptedRandomSessionKey, err = hex.DecodeString("7452ca55c225a1ca04b48fae32cf56fc")
	//	err = n.computeExportedSessionKey()
	//	checkV1Value(t, "ExportedSessionKey - NTLMSSP_REQUEST_NON_NT_SESSION_KEY", n.exportedSessionKey, "55555555555555555555555555555555", err)
	n.NegotiateFlags = NTLMSSP_REQUEST_NON_NT_SESSION_KEY.Unset(n.NegotiateFlags)

	// NTLMSSP_NEGOTIATE_LM_KEY is set:
	n.NegotiateFlags = NTLMSSP_NEGOTIATE_LM_KEY.Set(n.NegotiateFlags)
	err = n.computeKeyExchangeKey()
	//	n.encryptedRandomSessionKey, err = hex.DecodeString("4cd7bb57d697ef9b549f02b8f9b37864")
	//	err = n.computeExportedSessionKey()
	//	checkV1Value(t, "ExportedSessionKey - NTLMSSP_NEGOTIATE_LM_KEY", n.exportedSessionKey, "55555555555555555555555555555555", err)
	n.NegotiateFlags = NTLMSSP_NEGOTIATE_LM_KEY.Unset(n.NegotiateFlags)

	// 4.2.2.3 Messages
	challengeMessageBytes, _ := hex.DecodeString("4e544c4d53535000020000000c000c003800000033820a820123456789abcdef00000000000000000000000000000000060070170000000f530065007200760065007200")
	challengeMessage, err := ParseChallengeMessage(challengeMessageBytes)
	if err == nil {
		challengeMessage.String()
	} else {
		t.Errorf("Could not parse challenge message: %s", err)
	}

	client := new(V1ClientSession)
	client.SetUserInfo("User", "Password", "Domain")
	err = client.ProcessChallengeMessage(challengeMessage)
	if err != nil {
		t.Errorf("Could not process challenge message: %s", err)
	}

	server := new(V1ServerSession)
	server.SetUserInfo("User", "Password", "Domain")
	authenticateMessageBytes, err := hex.DecodeString("4e544c4d5353500003000000180018006c00000018001800840000000c000c00480000000800080054000000100010005c000000100010009c000000358280e20501280a0000000f44006f006d00610069006e00550073006500720043004f004d005000550054004500520098def7b87f88aa5dafe2df779688a172def11c7d5ccdef1367c43011f30298a2ad35ece64f16331c44bdbed927841f94518822b1b3f350c8958682ecbb3e3cb7")
	authenticateMessage, err := ParseAuthenticateMessage(authenticateMessageBytes, 1)
	if err == nil {
		authenticateMessage.String()
	} else {
		t.Errorf("Could not parse authenticate message: %s", err)
	}

	server = new(V1ServerSession)
	server.SetUserInfo("User", "Password", "Domain")
	server.serverChallenge = challengeMessage.ServerChallenge

	err = server.ProcessAuthenticateMessage(authenticateMessage)
	if err != nil {
		t.Errorf("Could not process authenticate message: %s", err)
	}
}

func TestNTLMv1WithClientChallenge(t *testing.T) {
	flags := uint32(0)
	flags = NTLMSSP_NEGOTIATE_56.Set(flags)
	flags = NTLMSSP_NEGOTIATE_VERSION.Set(flags)
	flags = NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY.Set(flags)
	flags = NTLMSSP_TARGET_TYPE_SERVER.Set(flags)
	flags = NTLMSSP_NEGOTIATE_ALWAYS_SIGN.Set(flags)
	flags = NTLMSSP_NEGOTIATE_NTLM.Set(flags)
	flags = NTLMSSP_NEGOTIATE_SEAL.Set(flags)
	flags = NTLMSSP_NEGOTIATE_SIGN.Set(flags)
	flags = NTLM_NEGOTIATE_OEM.Set(flags)
	flags = NTLMSSP_NEGOTIATE_UNICODE.Set(flags)

	n := new(V1Session)
	n.NegotiateFlags = flags
	n.responseKeyNT, _ = hex.DecodeString("a4f49c406510bdcab6824ee7c30fd852")
	n.responseKeyLM, _ = hex.DecodeString("e52cac67419a9a224a3b108f3fa6cb6d")
	n.clientChallenge, _ = hex.DecodeString("aaaaaaaaaaaaaaaa")
	n.serverChallenge, _ = hex.DecodeString("0123456789abcdef")

	var err error
	// 4.2.2.1.3 Session Base Key and Key Exchange Key
	err = n.computeExpectedResponses()
	err = n.computeSessionBaseKey()
	checkV1Value(t, "sessionBaseKey", n.sessionBaseKey, "d87262b0cde4b1cb7499becccdf10784", err)
	checkV1Value(t, "LMv1Response", n.lmChallengeResponse, "aaaaaaaaaaaaaaaa00000000000000000000000000000000", err)
	checkV1Value(t, "NTLMv1Response", n.ntChallengeResponse, "7537f803ae367128ca458204bde7caf81e97ed2683267232", err)
	err = n.computeKeyExchangeKey()
	checkV1Value(t, "keyExchangeKey", n.keyExchangeKey, "eb93429a8bd952f8b89c55b87f475edc", err)

	challengeMessageBytes, _ := hex.DecodeString("4e544c4d53535000020000000c000c003800000033820a820123456789abcdef00000000000000000000000000000000060070170000000f530065007200760065007200")
	challengeMessage, err := ParseChallengeMessage(challengeMessageBytes)
	if err == nil {
		challengeMessage.String()
	} else {
		t.Errorf("Could not parse challenge message: %s", err)
	}

	client := new(V1ClientSession)
	client.SetUserInfo("User", "Password", "Domain")
	err = client.ProcessChallengeMessage(challengeMessage)
	if err != nil {
		t.Errorf("Could not process challenge message: %s", err)
	}

	server := new(V1ServerSession)
	server.SetUserInfo("User", "Password", "Domain")
	server.serverChallenge = challengeMessage.ServerChallenge

	authenticateMessageBytes, _ := hex.DecodeString("4e544c4d5353500003000000180018006c00000018001800840000000c000c00480000000800080054000000100010005c000000000000009c000000358208820501280a0000000f44006f006d00610069006e00550073006500720043004f004d0050005500540045005200aaaaaaaaaaaaaaaa000000000000000000000000000000007537f803ae367128ca458204bde7caf81e97ed2683267232")
	authenticateMessage, err := ParseAuthenticateMessage(authenticateMessageBytes, 1)
	if err == nil {
		authenticateMessage.String()
	} else {
		t.Errorf("Could not parse authenticate message: %s", err)
	}

	err = server.ProcessAuthenticateMessage(authenticateMessage)
	if err != nil {
		t.Errorf("Could not process authenticate message: %s", err)
	}

	checkV1Value(t, "SealKey", server.ClientSealingKey, "04dd7f014d8504d265a25cc86a3a7c06", nil)
	checkV1Value(t, "SignKey", server.ClientSigningKey, "60e799be5c72fc92922ae8ebe961fb8d", nil)
}
