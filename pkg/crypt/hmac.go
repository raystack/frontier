package crypt

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
)

// GenerateHMAC produces a SHA-512/256 hash using 256-bit key
func GenerateHMAC(data []byte, key []byte) []byte {
	h := hmac.New(sha512.New512_256, key)
	h.Write(data)
	return h.Sum(nil)
}

// GenerateHMACFromHex produces a SHA-512/256 hash using 256-bit key
func GenerateHMACFromHex(data []byte, hexKey string) (string, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return "", err
	}
	result := GenerateHMAC(data, key)
	return hex.EncodeToString(result), nil
}

// VerifyHMAC checks the supplied MAC against the hash of the data using the key
func VerifyHMAC(data []byte, key []byte, suppliedMAC []byte) bool {
	expectedMAC := GenerateHMAC(data, key)
	return hmac.Equal(expectedMAC, suppliedMAC)
}

// VerifyHMACFromHex checks the supplied MAC against the hash of the data using the key
func VerifyHMACFromHex(data []byte, hexKey string, suppliedHexMAC string) (bool, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return false, err
	}
	suppliedMAC, err := hex.DecodeString(suppliedHexMAC)
	if err != nil {
		return false, err
	}
	return VerifyHMAC(data, key, suppliedMAC), nil
}
