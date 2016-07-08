//Copyright 2013 Thomson Reuters Global Resources. BSD License please see License file for more information

package ntlm

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestUTf16ToString(t *testing.T) {
	expected, _ := hex.DecodeString("5500730065007200")
	result := utf16FromString("User")
	if !bytes.Equal(expected, result) {
		t.Errorf("UTF16ToString failed got %s expected %s", hex.EncodeToString(result), "5500730065007200")
	}
}

func TestMacsEquals(t *testing.T) {
	// the MacsEqual should ignore the values in the second 4 bytes
	firstSlice := []byte{0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0xf8, 0xf9, 0xf0, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff}
	secondSlice := []byte{0xf1, 0xf2, 0xf3, 0xf4, 0x00, 0x00, 0x00, 0x00, 0xf9, 0xf0, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff}
	if !MacsEqual(firstSlice, secondSlice) {
		t.Errorf("Expected MacsEqual(%v, %v) to be true", firstSlice, secondSlice)
	}
}

func TestMacsEqualsFail(t *testing.T) {
	// the last bytes in the following test case should cause MacsEqual to return false
	firstSlice := []byte{0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0xf8, 0xf9, 0xf0, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff}
	secondSlice := []byte{0xf1, 0xf2, 0xf3, 0xf4, 0x00, 0x00, 0x00, 0x00, 0xf9, 0xf0, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xfe}
	if MacsEqual(firstSlice, secondSlice) {
		t.Errorf("Expected MacsEqual(%v, %v) to be false", firstSlice, secondSlice)
	}
}
