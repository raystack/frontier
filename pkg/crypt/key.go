package crypt

import (
	"crypto/rand"
	"encoding/hex"
	"io"
)

// NewEncryptionKey generates a random 256-bit key
func NewEncryptionKey() (*[32]byte, error) {
	key := [32]byte{}
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func NewEncryptionKeyInHex() (string, error) {
	key, err := NewEncryptionKey()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key[:]), nil
}
