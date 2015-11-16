//Copyright 2013 Thomson Reuters Global Resources. BSD License please see License file for more information

package ntlm

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func checkSigValue(t *testing.T, name string, value []byte, expected string, err error) {
	if err != nil {
		t.Errorf("Signature %s received error: %s", name, err)
	} else {
		expectedBytes, _ := hex.DecodeString(expected)
		if !bytes.Equal(expectedBytes, value) {
			t.Errorf("Signature %s is not correct got %s expected %s", name, hex.EncodeToString(value), expected)
		}
	}
}

// 4.2.2.4 GSS_WrapEx Examples
func TestSealWithoutExtendedSessionSecurity(t *testing.T) {
	key, _ := hex.DecodeString("55555555555555555555555555555555")
	handle, _ := rc4Init(key)
	plaintext, _ := hex.DecodeString("50006c00610069006e007400650078007400")
	seqNum := uint32(0)
	flags := uint32(0)

	sealed, sig := seal(flags, handle, nil, seqNum, plaintext)
	checkSigValue(t, "Sealed message", sealed, "56fe04d861f9319af0d7238a2e3b4d457fb8", nil)
	checkSigValue(t, "Randompad", sig.RandomPad, "00000000", nil)
	checkSigValue(t, "RC4 Checksum", sig.CheckSum, "09dcd1df", nil)
	checkSigValue(t, "Xor Seq", sig.SeqNum, "2e459d36", nil)
}

func TestSealSignWithExtendedSessionSecurity(t *testing.T) {
	sealKey, _ := hex.DecodeString("04dd7f014d8504d265a25cc86a3a7c06")
	signKey, _ := hex.DecodeString("60e799be5c72fc92922ae8ebe961fb8d")
	handle, _ := rc4Init(sealKey)
	plaintext, _ := hex.DecodeString("50006c00610069006e007400650078007400")
	seqNum := uint32(0)
	flags := uint32(0)
	flags = NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY.Set(flags)

	sealed, sig := seal(flags, handle, signKey, seqNum, plaintext)
	checkSigValue(t, "Sealed Data", sealed, "a02372f6530273f3aa1eb90190ce5200c99d", nil)
	checkSigValue(t, "CheckSum", sig.CheckSum, "ff2aeb52f681793a", nil)
	checkSigValue(t, "Signature", sig.Bytes(), "01000000ff2aeb52f681793a00000000", nil)
}

func TestSealSignWithExtendedSessionSecurityKeyEx(t *testing.T) {
	sealKey, _ := hex.DecodeString("59f600973cc4960a25480a7c196e4c58")
	signKey, _ := hex.DecodeString("4788dc861b4782f35d43fd98fe1a2d39")
	handle, _ := rc4Init(sealKey)
	plaintext, _ := hex.DecodeString("50006c00610069006e007400650078007400")
	seqNum := uint32(0)
	flags := uint32(0)
	flags = NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY.Set(flags)
	flags = NTLMSSP_NEGOTIATE_KEY_EXCH.Set(flags)

	sealed, sig := seal(flags, handle, signKey, seqNum, plaintext)
	checkSigValue(t, "Sealed Data", sealed, "54e50165bf1936dc996020c1811b0f06fb5f", nil)
	checkSigValue(t, "RC4 CheckSum", sig.CheckSum, "7fb38ec5c55d4976", nil)
	checkSigValue(t, "Signature", sig.Bytes(), "010000007fb38ec5c55d497600000000", nil)
}
