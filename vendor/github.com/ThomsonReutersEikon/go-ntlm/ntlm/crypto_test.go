//Copyright 2013 Thomson Reuters Global Resources. BSD License please see License file for more information

package ntlm

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestMd4(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	byteData, _ := hex.DecodeString("93ebafdfedd1994e8018cc295cc1a8ee")
	if !bytes.Equal(md4(data), byteData) {
		t.Error("MD4 result not correct")
	}
}

func TestHmacMd5(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	byteData, _ := hex.DecodeString("9155578efbf3810a2adb4dee232a5fee")
	if !bytes.Equal(hmacMd5(data, data), byteData) {
		t.Error("HmacMd5 result not correct")
	}
}

func TestNonce(t *testing.T) {
	data := nonce(10)
	if len(data) != 10 {
		t.Error("Nonce is incorrect length")
	}
}

func TestRc4K(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	key := []byte{1, 2, 3, 4, 5}
	result, err := rc4K(key, data)
	if err != nil {
		// TODO: Need some sample data to test RC4K
		//	t.Error("Error returned for RC4K")
	}
	if !bytes.Equal(result, data) {
		//	t.Error("RC4K result not correct")
	}
}

func TestDesL(t *testing.T) {
	key, _ := hex.DecodeString("e52cac67419a9a224a3b108f3fa6cb6d")
	message := []byte("12345678")
	result, _ := desL(key, message)
	expected, _ := hex.DecodeString("1192855D461A9754D189D8AE94D82488E3707C0662C0476A")
	if !bytes.Equal(result, expected) {
		t.Errorf("DesL did not produce correct result, got %s expected %s", hex.EncodeToString(result), hex.EncodeToString(expected))
	}
}

func TestCRC32(t *testing.T) {
	bytes := []byte("Discard medicine more than two years old.")
	result := crc32(bytes)
	expected := uint32(0x6b9cdfe7)
	if expected != result {
		t.Errorf("CRC 32 data is not correct got %d expected %d", result, expected)
	}
}
