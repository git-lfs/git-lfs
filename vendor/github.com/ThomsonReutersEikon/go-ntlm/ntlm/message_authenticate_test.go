//Copyright 2013 Thomson Reuters Global Resources. BSD License please see License file for more information

package ntlm

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"testing"
)

func checkPayloadStruct(t *testing.T, payloadStruct *PayloadStruct, len uint16, offset uint32) {
	if payloadStruct.Len != len || payloadStruct.Offset != offset {
		t.Errorf("Failed to parse payload struct %d, %d", payloadStruct.Len, payloadStruct.Offset)
	}
}

func TestParseNTLMv1AsV2(t *testing.T) {
	ntlmv1data := "TlRMTVNTUAADAAAAGAAYALYAAAAYABgAzgAAADQANABIAAAAIAAgAHwAAAAaABoAnAAAABAAEADmAAAAVYKQQgUCzg4AAAAPYQByAHIAYQB5ADEAMgAuAG0AcwBnAHQAcwB0AC4AcgBlAHUAdABlAHIAcwAuAGMAbwBtAHUAcwBlAHIAcwB0AHIAZQBzAHMAMQAwADAAMAAwADgATgBZAEMAVgBBADEAMgBTADIAQwBNAFMAQQDguXWdC2hLH+C5dZ0LaEsf4Ll1nQtoSx9nI+fkE73qtElnkDiSQbxfcDN9zbtO1qfyK3ZTI6CUhvjxmXnpZEjY"
	authBytes, err := base64.StdEncoding.DecodeString(ntlmv1data)
	_, err = ParseAuthenticateMessage(authBytes, 2)
	if err == nil {
		t.Error("Should have returned error when tring to parse an NTLMv1 authenticate message as NTLMv2")
	}
	_, err = ParseAuthenticateMessage(authBytes, 1)
	if err != nil {
		t.Error("Should not have returned error when tring to parse an NTLMv1 authenticate message")
	}
}

func TestAuthenticateNtlmV1(t *testing.T) {
	authenticateMessage := "TlRMTVNTUAADAAAAGAAYAIgAAAAYABgAoAAAAAAAAABYAAAAIAAgAFgAAAAQABAAeAAAABAAEAC4AAAAVYKQYgYBsR0AAAAP2BgW++b14Dh6Z5B4Xs1DiHAAYQB1AGwAQABwAGEAdQBsAGQAaQB4AC4AbgBlAHQAVwBJAE4ANwBfAEkARQA4ACugxZFzvHB4P6LdKbbZpiYHo2ErZURLiSugxZFzvHB4P6LdKbbZpiYHo2ErZURLibmpCUlnbq2I4LAdEhLdg7I="
	authenticateData, err := base64.StdEncoding.DecodeString(authenticateMessage)

	if err != nil {
		t.Error("Could not base64 decode message data")
	}

	a, err := ParseAuthenticateMessage(authenticateData, 1)
	if err != nil {
		t.Error("Could not parse authenticate message")
	}

	a.String()

	outBytes := a.Bytes()

	if len(outBytes) > 0 {
		reparsed, err := ParseAuthenticateMessage(outBytes, 1)
		if err != nil {
			t.Error("Could not re-parse authenticate message")
		}
		if reparsed.String() != a.String() {
			t.Error("Reparsed message is not the same")
		}
	} else {
		t.Error("Invalid authenticate messsage bytes")
	}
}

func TestAuthenticateNtlmV2(t *testing.T) {
	authenticateMessage := "TlRMTVNTUAADAAAAGAAYAI4AAAAGAQYBpgAAAAAAAABYAAAAIAAgAFgAAAAWABYAeAAAABAAEACsAQAAVYKQQgYAchcAAAAPpdhi9ItaLWwSGpFMT4VQbnAAYQB1AGwAQABwAGEAdQBsAGQAaQB4AC4AbgBlAHQASQBQAC0AMABBADAAQwAzAEEAMQBFAAE/QEbbIB1InAX5KMgp4s4wmpPZ9jp9T3EC95rRY01DhMSv1kei5wYBAQAAAAAAADM6xfahoM0BMJqT2fY6fU8AAAAAAgAOAFIARQBVAFQARQBSAFMAAQAcAFUASwBCAFAALQBDAEIAVABSAE0ARgBFADAANgAEABYAUgBlAHUAdABlAHIAcwAuAG4AZQB0AAMANAB1AGsAYgBwAC0AYwBiAHQAcgBtAGYAZQAwADYALgBSAGUAdQB0AGUAcgBzAC4AbgBlAHQABQAWAFIAZQB1AHQAZQByAHMALgBuAGUAdAAIADAAMAAAAAAAAAAAAAAAADAAAFaspfI82pMCKSuN2L09orn37EQVvxCSqVqQhCloFhQeAAAAAAAAAADRgm1iKYwwmIF3axms/dIe"
	authenticateData, err := base64.StdEncoding.DecodeString(authenticateMessage)

	if err != nil {
		t.Error("Could not base64 decode message data")
	}

	a, err := ParseAuthenticateMessage(authenticateData, 2)

	if err != nil || a == nil {
		t.Error("Failed to parse authenticate message " + err.Error())
	}

	checkPayloadStruct(t, a.LmChallengeResponse, 24, 142)
	checkPayloadStruct(t, a.NtChallengeResponseFields, 262, 166)
	checkPayloadStruct(t, a.DomainName, 0, 88)
	checkPayloadStruct(t, a.UserName, 32, 88)
	checkPayloadStruct(t, a.Workstation, 22, 120)
	checkPayloadStruct(t, a.EncryptedRandomSessionKey, 16, 428)

	if a.NegotiateFlags != uint32(1116766805) {
		t.Errorf("Authenticate negotiate flags not correct should be %d got %d", uint32(1116766805), a.NegotiateFlags)
	}

	mic, err := hex.DecodeString("a5d862f48b5a2d6c121a914c4f85506e")
	if !bytes.Equal(a.Mic, mic) {
		t.Errorf("Mic not correct, should be %s, got %s", "a5d862f48b5a2d6c121a914c4f85506e", hex.EncodeToString(a.Mic))
	}

	if len(a.Payload) != 356 {
		t.Errorf("Length of payload is incorrect got: %d, should be %d", len(a.Payload), 356)
	}

	a.String()

	// Generate the bytes from the message and reparse it and make sure that works
	bytes := a.Bytes()
	if len(bytes) == 0 {

	}
}
